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
