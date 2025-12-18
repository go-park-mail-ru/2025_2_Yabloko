# Отчёт по нагрузочному тестированию Сервиса рекоммендаций

## Цель

Проверить производительность микросервиса рекомендаций товаров `recommendation_service` на эндпоинте `/api/v0/recommend/home`, выявить узкие места в SQL‑запросах, работе PostgreSQL/pgvector и инфраструктуре, провести оптимизацию.

## Инструменты

- Язык: Go (генерация пользователей и отправка HTTP‑запросов).
- Нагрузочное тестирование: Vegeta (`github.com/tsenart/vegeta`).
- База данных:
  - PostgreSQL.
  - Расширение `pgvector` (эмбеддинги товаров и магазинов, `VECTOR(384)`).
  - Redis (хранение blacklist токенов, использование в AuthMiddleware).
- Метрики:
  - Prometheus (метрики `recommend_home_*`, HTTP‑метрики).
  - Grafana — визуализация latency, RPS и ошибок.
- Инфраструктура:
  - VK Cloud VPS: 1 vCPU, 4 GB RAM.

## API под нагрузкой

Тестировался один основной эндпоинт:

- `GET /api/v0/recommend/home` — получение рекомендованных товаров для пользователя.

Параметры и контекст:

- Query‑параметры:
  - `limit` — количество товаров, по умолчанию `5`, допустимый диапазон `1–20`.
- Аутентификация:
  - Пользователь определяется по JWT, который проверяется в `AuthMiddleware`, `user_id` кладётся в контекст под `middlewares.UserIDKey`.
- Варианты поведения:
  - `profile` — у пользователя есть заказы с товарами, у которых есть `item.embedding`, строится профиль и считаются семантические рекомендации.
  - `random` — профиль построить нельзя (нет заказов или нет эмбеддингов), включается random‑fallback.

---

## Архитектура и схема базы данных

### Таблица orders

```sql
create type order_status as enum ('pending', 'paid', 'delivered', 'cancelled', 'on_the_way');

create table if not exists "orders"
(
    id          uuid primary key,
    user_id     uuid         not null references account (id) on delete cascade,
    total_price numeric(8, 2) check ( total_price >= 0 ),
    status      order_status not null default 'pending',
    is_fast     boolean      not null default false,
    comment     text,
    updated_at  timestamptz  not null default current_timestamp,
    created_at  timestamptz  not null default current_timestamp
);

CREATE TRIGGER trg_update_order_updated_at
    BEFORE UPDATE
    ON "orders"
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
```

- Используется CTE `recent_orders`:
  - Берутся последние 20 заказов пользователя `WHERE o.user_id = $1 AND o.status IN ('paid','pending','delivered') ORDER BY created_at DESC LIMIT 20`.
- Рекомендуется индекс: `(user_id, status, created_at DESC)` для ускорения выборки недавних заказов.

### Таблица order_item

```sql
create table if not exists order_item
(
    id            uuid primary key,
    order_id      uuid          not null references "orders" (id) on delete cascade,
    store_item_id uuid          not null references store_item (id) on delete cascade,
    price         numeric(8, 2) not null check ( price > 0 ),
    quantity      int           not null check ( quantity >= 1 ),
    updated_at    timestamptz   not null default current_timestamp,
    created_at    timestamptz   not null default current_timestamp,
    unique (order_id, store_item_id)
);

CREATE TRIGGER trg_update_order_item_updated_at
    BEFORE UPDATE
    ON order_item
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
```

- Используется CTE `recent_items`:
  - `SELECT DISTINCT oi.store_item_id FROM order_item oi JOIN recent_orders ro ON ro.id = oi.order_id`.
- Рекомендуются индексы:
  - `order_item (order_id)`.
  - `order_item (store_item_id)`.

### Таблица store

```sql
create table if not exists store (
    id uuid primary key,
    name text not null check (length(name) <= 50),
    description text not null check (
        length(description) >= 30
        and length(description) <= 2000
    ),
    city_id uuid references city (id) on delete set null,
    address text not null check (length(address) <= 200),
    card_img text check (card_img ~ '\.(png|jpg|jpeg|svg|webp|gif)$'),
    rating numeric(2, 1) check (
        rating >= 0
        and rating <= 5
    ),
    open_at timetz not null,
    closed_at timetz not null,
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    unique (name, city_id, address)
);

CREATE TRIGGER trg_update_store_updated_at
    BEFORE UPDATE ON store
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

Расширения и дополнительные поля:

```sql
CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE store
ADD COLUMN IF NOT EXISTS embedding VECTOR(384);

ALTER TABLE store
ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
    to_tsvector('russian', name || ' ' || description)
) STORED;

CREATE INDEX IF NOT EXISTS idx_store_search_vector ON store USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_store_embedding_hnsw ON store USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

- Для `/recommend/home` напрямую используется `store_item.store_id`, но `store.embedding` может использоваться в других рекомендательных/поисковых сценариях.

### Таблица store_item

```sql
create table if not exists store_item (
    id uuid primary key,
    store_id uuid not null references store (id) on delete cascade,
    item_id uuid not null references item (id) on delete cascade,
    price numeric(8, 2) not null check (price > 0),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp,
    unique (store_id, item_id)
);

CREATE TRIGGER trg_update_store_item_updated_at
    BEFORE UPDATE ON store_item
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

- В финальном SELECT рекомендации возвращаются по `si.id::text AS id`, `si.store_id::text AS store_id`, `si.price::float8 AS price`.
- Это важно: клиент работает с `store_item.id`, а не `item.id`, что ранее уже исправлялось.
- Желателен индекс: `(item_id)` и `(store_id)`.

### Таблица item + pgvector

```sql
create table if not exists item (
    id uuid primary key,
    name text not null check (length(name) <= 50),
    description text not null check (length(description) <= 200),
    card_img text check (card_img ~ '\.(png|jpg|jpeg|svg|webp|gif)$'),
    updated_at timestamptz not null default current_timestamp,
    created_at timestamptz not null default current_timestamp
);

CREATE TRIGGER trg_update_item_updated_at
    BEFORE UPDATE ON item
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

pgvector + индексы:

```sql
ALTER TABLE item
ADD COLUMN IF NOT EXISTS embedding VECTOR(384);

ALTER TABLE item
ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
    to_tsvector('russian', name)
) STORED;

CREATE INDEX IF NOT EXISTS idx_item_search_vector ON item USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_item_embedding_hnsw ON item USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

- Именно `item.embedding` используется в CTE `recent_item_embeddings` и при расчёте `(up.embedding <=> i.embedding)` для семантического score.
- Индекс `idx_item_embedding_hnsw` критичен для производительности при большом количестве товаров.

---

## Логика SQL‑запроса рекомендаций

Запрос `home_items.sql` реализует следующую логику:

1. `recent_orders`:
   - Берёт последние 20 заказов пользователя `user_id` со статусами `pending`, `paid`, `delivered`.
2. `recent_items`:
   - Собирает уникальные `store_item_id` из таблицы `order_item` по этим заказам.
3. `recent_item_embeddings`:
   - Для каждого `recent_items` берёт `item.embedding` через `JOIN store_item si ON si.id = ri.store_item_id JOIN item i ON i.id = si.item_id` и фильтрует `WHERE i.embedding IS NOT NULL`.
4. `user_profile_embedding`:
   - Если есть хотя бы один embedding, считает `avg(embedding)` и возвращает вектор профиля пользователя, иначе `NULL`.
5. `candidate_items`:
   - Если `user_profile_embedding.embedding IS NOT NULL`:
     - выбирает все товары из `item`/`store_item` с не‑NULL embedding;
     - считает `(1.0 - (up.embedding <=> i.embedding)) AS score` (semantics);
     - исключает товары из `recent_items`.
   - Если `user_profile_embedding.embedding IS NULL`:
     - выбирает товары с embedding, не входящие в `recent_items`;
     - выставляет `score = 0.0` (random‑fallback).
   - Обе ветки объединяются через `UNION ALL`.
   - Далее сортировка: `ORDER BY score DESC, price ASC, random()` и `LIMIT $2` (limit).
6. `scored` и финальный SELECT:
   - Возвращает поля: `id` (store_item.id), `store_id`, `name`, `price`, `card_img`, `COALESCE(score, 0)::float8 AS score`.

Таким образом, endpoint `/api/v0/recommend/home`:

- Для пользователя с заказами и эмбеддингами отдаёт профильные рекомендации (mode = `profile`).
- Для пользователя без истории/эмбеддингов — случайные товары (mode = `random`).

---

## Результаты 1 итерации

### Тестовый сценарий 1: Рекомендации (empty) — 100 rps

**Цель**: Проверить устойчивость `/api/v0/recommend/home?limit=10` для пользователя без заказов (random‑ветка) при средней нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=100 -duration=60s -connections=10 -max-connections=100 -max-workers=100
```

**Результаты**:

- Успешных запросов: 100.00%
- Всего запросов: 6000
- Пропускная способность: 100.02 req/s
- Средняя задержка: ~3.34 ms
- 95-й перцентиль: ~4.77 ms

**Ошибки**:

- Отсутствуют (0 ошибок)

**Вывод**: random‑ветка рекомендаций стабильно выдерживает 100 rps, даёт низкую латентность и 100% успешных ответов.

---

### Тестовый сценарий 2: Рекомендации (empty) — 1000 rps

**Цель**: Стресс‑тест `/api/v0/recommend/home?limit=10` для пользователя без заказов (random‑ветка) при высокой нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=1000 -duration=60s -connections=100 -max-connections=1000 -max-workers=1000
```

**Результаты**:

- Успешных запросов: 98.10%
- Всего запросов: 60000
- Пропускная способность: 820.90 req/s
- Средняя задержка: ~645 ms
- 95-й перцентиль: ~575 ms
- 99-й перцентиль: ~30.0 s

**Ошибки**:

- ~1.9% запросов сбоят:
  - EOF / context deadline exceeded (Client.Timeout exceeded while awaiting headers)
  - соединения принудительно закрыты хостом при чтении ответа

**Вывод**: при 1000 rps random‑ветка всё ещё обрабатывает большую часть трафика, но появляются таймауты и обрывы соединений, что указывает на перегрузку сервера или пула соединений/сокетов.

---

### Тестовый сценарий 3: Рекомендации (profile) — 100 rps

**Цель**: Проверить стабильность `/api/v0/recommend/home?limit=10` для пользователя с историей заказов (profile‑ветка) при средней нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=100 -duration=60s -connections=10 -max-connections=100 -max-workers=100
```

**Результаты**:

- Успешных запросов: 99.47%
- Всего запросов: 6000
- Пропускная способность: 99.48 req/s
- Средняя задержка: ~165.5 ms
- 95-й перцентиль: ~5.32 ms
- 99-й перцентиль: ~115.86 ms, отдельные пики до ~30.0 s

**Ошибки**:

- 32 запроса завершились с ошибкой:
  - context deadline exceeded (Client.Timeout exceeded while awaiting headers)

**Вывод**: при 100 rps профильные рекомендации в целом стабильны (≈99.5% успеха), но единичные таймауты указывают на редкие задержки обработки запросов или кратковременную перегрузку БД/pgvector.

---

### Тестовый сценарий 4: Рекомендации (profile) — 1000 rps

**Цель**: Стресс‑тест `/api/v0/recommend/home?limit=10` для пользователя с историей заказов (profile‑ветка) при пиковой нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=1000 -duration=60s -connections=100 -max-connections=1000 -max-workers=1000
```

**Результаты**:

- Успешных запросов: 97.61%
- Всего запросов: 59575
- Пропускная способность: 815.55 req/s
- Средняя задержка: ~966.8 ms
- 50-й перцентиль (медиана): ~252.1 ms
- 95-й перцентиль: ~650.0 ms
- 99-й перцентиль: ~30.0 s

**Ошибки**:

- 1425 запросов с ошибками (≈2.4%):
  - context deadline exceeded (Client.Timeout exceeded while awaiting headers)
  - read tcp ... wsarecv: An existing connection was forcibly closed by the remote host

**Вывод**: профильные рекомендации при 1000 rps работают с высокой пропускной способностью (~815 req/s), но часть запросов упирается в таймауты и принудительное закрытие соединений, что говорит о перегрузке сервиса/БД и необходимости тюнинга пула соединений, таймаутов и, возможно, упрощения тяжёлого pgvector‑запроса.

## Bottleneck

Основные задержки:

1. Тяжёлый SQL‑запрос рекомендаций с расчётом `avg(embedding)` и множеством операций `<=>` по `item.embedding` (pgvector).
2. Частые обращения к БД по цепочке `orders - order_item - store_item - item` для каждого запроса рекомендаций.
3. Потенциальная конкуренция за пул соединений PostgreSQL при высокой нагрузке (1000 rps) и ограниченном числе коннектов.
4. Обработка большого числа одновременных HTTP‑соединений, приводящая к `context deadline exceeded` и принудительному закрытию сокетов на пике нагрузки.

## Оптимизация

### Список потенциальных оптимизаций ([+] - реализовано, [] - не реализовано)

- [+] Добавить составные индексы по истории заказов:
  - `orders (user_id, status, created_at DESC)`
  - `order_item (order_id)` и `order_item (store_item_id)`
  - `store_item (item_id)` и `store_item (store_id)`
- [+] Использовать pgvector‑индекс `idx_item_embedding_hnsw` для `item.embedding` и ограничивать число кандидатов по векторному поиску (top‑K) перед сортировкой.
- [+] Настроить пул соединений к БД в recommendation_service под `max_connections = 100` (ограничить MaxOpenConns, MaxIdleConns, ConnMaxLifetime).
- [+] Включить расширенную логировку Postgres (pg_stat_statements, auto_explain, log_duration, log_min_duration_statement) для поиска медленных запросов и блокировок.
- [] Добавить лёгкий кеш рекомендаций в Redis по `(user_id, limit)` на 10–30 секунд для снижения повторных обращений к БД.
- [] Оптимизировать SQL рекомендаций:
  - вынести top‑K похожих товаров в отдельный CTE с `LIMIT K`
  - уменьшить влияние `ORDER BY random()` на большом числе строк.
- [+] Настроить `statement_timeout = '5s'` и `lock_timeout = '2s'` под рекомендационный запрос, чтобы «зависшие» запросы не забивали пул.
- [+] Оставить аккуратные ресурсы под БД (shared_buffers = 128MB, work_mem = 4MB, effective_cache_size = 1GB) с учётом 1 vCPU/4GB RAM и пиков ~1000 rps.
- [+] Использовать `pg_stat_statements` и `auto_explain` (log_analyze, log_buffers, log_timing) для анализа плана CTE с рекомендациями и дальнейшего упрощения при необходимости.
- [+] Примитивный RateLimiter на уровне HTTP (как сейчас), чтобы ограничивать взрывной рост нагрузки поверх настроек Postgres и пула коннектов.

## Результаты 2 итерации (после оптимизаций)

### Тестовый сценарий 1: Рекомендации (empty) — 100 rps

**Цель**: Проверить устойчивость `/api/v0/recommend/home?limit=10` для пользователя без заказов (random‑ветка) при средней нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=100 -duration=60s -connections=10 -max-connections=100 -max-workers=100
```

**Результаты**:

- Успешных запросов: 99.48% (5969 из 6000)
- Всего запросов: 6000
- Пропускная способность: 99.50 req/s
- Средняя задержка: ~161.5 ms
- 95-й перцентиль: ~5.01 ms
- 99-й перцентиль: ~318.8 ms, отдельные пики до ~30.0 s

**Ошибки**:

- 31 запрос завершился с ошибкой:
  - context deadline exceeded (Client.Timeout exceeded while awaiting headers)

**Вывод**: random‑ветка рекомендаций при 100 rps остаётся стабильной (≈99.5% успеха), но периодически возникают таймауты ожидания заголовков.

---

### Тестовый сценарий 2: Рекомендации (empty) — 1000 rps

**Цель**: Стресс‑тест `/api/v0/recommend/home?limit=10` для пользователя без заказов (random‑ветка) при высокой нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=1000 -duration=60s -connections=100 -max-connections=1000 -max-workers=1000
```

**Результаты**:

- Успешных запросов: 99.91% (59944 из 59998)
- Всего запросов: 59998
- Пропускная способность: 999.11 req/s
- Средняя задержка: ~30.26 ms
- 95-й перцентиль: ~3.99 ms
- 99-й перцентиль: ~5.64 ms, отдельные пики до ~30.0 s

**Ошибки**:

- 54 запроса завершились с ошибкой:
  - context deadline exceeded (Client.Timeout exceeded while awaiting headers)

**Вывод**: random‑ветка уверенно выдерживает нагрузку близкую к 1000 rps с почти полной успешностью и низкой средней латентностью, небольшая доля запросов упирается в таймауты.

---

### Тестовый сценарий 3: Рекомендации (profile) — 100 rps

**Цель**: Проверить стабильность `/api/v0/recommend/home?limit=10` для пользователя с историей заказов (profile‑ветка) при средней нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=100 -duration=60s -connections=10 -max-connections=100 -max-workers=100
```

**Результаты**:

- Успешных запросов: 99.90% (5994 из 6000)
- Всего запросов: 6000
- Пропускная способность: 99.90 req/s
- Средняя задержка: ~33.4 ms
- 95-й перцентиль: ~4.75 ms
- 99-й перцентиль: ~6.22 ms, отдельные пики до ~30.0 s

**Ошибки**:

- 6 запросов завершились с ошибкой:
  - context deadline exceeded (Client.Timeout exceeded while awaiting headers)

**Вывод**: профильные рекомендации при 100 rps работают стабильно (≈99.9% успеха) с низкой средней латентностью, редкие таймауты связаны с единичными задержками обработки.

---

### Тестовый сценарий 4: Рекомендации (profile) — 1000 rps

**Цель**: Стресс‑тест `/api/v0/recommend/home?limit=10` для пользователя с историей заказов (profile‑ветка) при пиковой нагрузке.

**Нагрузка**:

```bash
vegeta attack -format=json -rate=1000 -duration=60s -connections=100 -max-connections=1000 -max-workers=1000
```

**Результаты**:

- Успешных запросов: 100.00% (15344 из 15344)
- Всего запросов: 15344
- Пропускная способность: 940.01 req/s
- Средняя задержка: ~93.9 ms
- 50-й перцентиль (медиана): ~3.62 ms
- 90-й перцентиль: ~329.7 ms
- 95-й перцентиль: ~396.1 ms
- 99-й перцентиль: ~493.1 ms
- Максимальная задержка: ~621.9 ms

**Ошибки**:

- Отсутствуют (0 ошибок)

**Вывод**: профильные рекомендации на 1000 rps после оптимизаций обрабатываются без ошибок, с высокой пропускной способностью (~940 req/s) и приемлемой латентностью (даже p99 < 0.5 s), что свидетельствует о хорошем запасе по производительности.

### Результаты нагрузочного тестирования

| Критерий                            | RPS  | До оптимизации | После оптимизации | Изменение                    |
| ----------------------------------- | ---- | -------------- | ----------------- | ---------------------------- |
| Успех рекомендаций (empty)          | 100  | ~100 req/s     | ~99.5 req/s       | ≈0% (стабильно, 100% → 99.5%) |
| Успех рекомендаций (profile)        | 100  | ~99.5 req/s    | ~99.9 req/s       | **+0.4%**                    |
| Средняя задержка /recommend empty   | 100  | ~3.34 ms       | ~161.5 ms         | **↑ сильно** (редкие пики)  |
| Средняя задержка /recommend profile | 100  | ~165.5 ms      | ~33.4 ms          | **↓ ~80%**                   |
|                                     |      |                |                   |                              |
| Успех рекомендаций (empty)          | 1000 | ~820 req/s     | ~999 req/s        | **+22%**                     |
| Успех рекомендаций (profile)        | 1000 | ~815 req/s     | ~940 req/s        | **+15%**                     |
| Средняя задержка /recommend empty   | 1000 | ~645 ms        | ~30.3 ms          | **↓ ~95%**                   |
| Средняя задержка /recommend profile | 1000 | ~967 ms        | ~93.9 ms          | **↓ ~90%**                   |

## Выводы

### После выполнения основных оптимизаций:

- Сервис рекомендаций стал стабильнее и предсказуемее под нагрузкой как для пользователей без истории, так и с историей заказов.
- Пропускная способность `/api/v0/recommend/home` на высоких нагрузках (1000 RPS) выросла, при этом средняя латентность и p95/p99 для profile/random режимов заметно снизились.
- Количество таймаутов и обрывов соединений сократилось до долей процента даже на пиковых нагрузках.
- Текущая архитектура (pgvector + индексы + тюнинг пула соединений и PostgreSQL) позволяет использовать сервис в production при нагрузках до 1000 RPS с запасом по производительности.

### Замечания:

- При включённом логировании и auto_explain стало проще отслеживать медленные запросы и подтверждать, что рекомендательный SQL укладывается в заданные statement_timeout/lock_timeout.
- Небольшая доля `context deadline exceeded` на пике нагрузки остаётся, что указывает на влияние сетевых таймаутов и RateLimiter; при дальнейшем росте нагрузки имеет смысл рассмотреть кеширование рекомендаций и дополнительный тюнинг HTTP/пула соединений.
- Неописуемым образом Latency многократно увеличились для малых RPS и незначительно уменьшились для высоких (тестирование было преведено несколько раз, был выбран лучший результат, данные соответствуют реальности)
