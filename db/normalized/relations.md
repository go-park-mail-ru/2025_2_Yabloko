# Описание таблиц

___

## account

    Информация о зарегистрированных пользователях

* id — id пользователя
* email — электронная почта пользователя
* password — хэш пароля
* name — имя пользователя
* phone — номер телефона в формате (+...)
* city_id — id города пользователя
* address — адрес по умолчанию для доставки

**Соотношение:**  
{id} → {email, password, name, phone, city_id, address}
{email} → {id, password, name, phone, city_id, address}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## store

    Информация о магазинах и ресторанах

* id — id магазина
* name — название
* description — краткое описание
* city_id — id города расположения магазина
* address — адрес магазина
* card_img — путь до изображения карточки магазина
* rating — средний рейтинг (0.0–5.0)
* open_at — время открытия
* closed_at — время закрытия

**Соотношение:**  
{id} → {name, description, city_id, address, card_img, rating, open_at, closed_at}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## city

    Справочник городов, в которых работают магазины и доставка

* id — id города
* name — название города

**Соотношение:**  
{id} → {name}
{name} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## friend

    Список друзей пользователей

* id — id записи
* account_1_id — id первого пользователя
* account_2_id — id второго пользователя

**Соотношение:**  
{id} → {account_1_id, account_2_id}
{account_1_id, account_2_id} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## cart

    Корзина пользователя

* id — id корзины
* user_id — id пользователя

**Соотношение:**  
{id} → {user_id}  
{user_id} → {id}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## cart_item

    Связь корзины с товарами

* id — id записи
* cart_id — id корзины
* store_item_id — id товара в магазине
* quantity — количество товара

**Соотношение:**  
{id} → {cart_id, store_item_id, quantity}  
{cart_id, store_item_id} → {id, quantity}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## order

    Информация о заказах

* id — id заказа
* user_id — id пользователя, сделавшего заказ
* store_id — id магазина, из которого сделан заказ
* total_price — сумма заказа
* status — статус заказа (pending, completed, cancelled)

**Соотношение:**  
{id} → {user_id, store_id, total_price, status}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## order_item

    Связь товаров и заказа

* id — id записи
* order_id — id заказа
* store_item_id — id товара в магазине
* price — цена товара на момент оформления заказа
* quantity — количество единиц товара

**Соотношение:**  
{id} → {order_id, store_item_id, price, quantity}
{order_id, store_item_id} → {id, price, quantity}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## tag

    Справочник тегов (например: самовывоз, Грузия, вок, пицца)

* id — id тега
* name — название тега

**Соотношение:**  
{id} → {name}  
{name} → {id}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## store_tag

    Связь тегов и магазинов

* id — id записи
* store_id — id магазина
* tag_id — id тега

**Соотношение:**  
{id} → {store_id, tag_id}  
{store_id, tag_id} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## category

    Справочник категорий магазинов (например: ресторан, аптека, продукты)

* id — id категории
* name — название категории

**Соотношение:**  
{id} → {name}  
{name} → {id}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## store_category

    Связь категорий и магазинов

* id — id записи
* store_id — id магазина
* category_id — id категории

**Соотношение:**  
{id} → {store_id, category_id}  
{store_id, category_id} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## item

    Справочник товаров

* id — id товара
* name — название
* description — описание
* card_img — путь к изображению карточки товара

**Соотношение:**  
{id} → {name, description, card_img}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## store_item

    Товары, представленные в конкретном магазине

* id — id записи
* store_id — id магазина
* item_id — id товара
* price — актуальная цена товара в магазине

**Соотношение:**  
{id} → {store_id, item_id, price}  
{store_id, item_id} → {id, price}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## type

    Справочник типов товаров (например: напитки, супы, гарниры)

* id — id типа
* name — название типа

**Соотношение:**  
{id} → {name}
{name} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## item_type

    Связь товаров и типов

* id — id записи
* item_id — id товара
* type_id — id типа

**Соотношение:**  
{id} → {item_id, type_id}
{item_id, type_id} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## promotion

    Акции и скидки

* id — id акции
* name — название акции
* relative_discount — относительная скидка (в процентах)
* absolute_discount — абсолютная скидка (в валюте)
* start_at — дата и время начала действия акции
* end_at — дата и время окончания действия акции

**Соотношение:**  
{id} → {name, relative_discount, absolute_discount, start_at, end_at}  
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## promocode

    Промокоды

* id — id промокода
* code — текст промокода
* relative_discount — относительная скидка
* absolute_discount — абсолютная скидка
* user_id — id пользователя (если промокод персональный)
* start_at — дата и время начала действия
* end_at — дата и время окончания действия

**Соотношение:**  
{id} → {code, relative_discount, absolute_discount, user_id, start_at, end_at}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## promotion_item

    Связь товаров и акций

* id — id записи
* item_id — id товара
* promotion_id — id акции

**Соотношение:**  
{id} → {item_id, promotion_id}
{item_id, promotion_id} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## promocode_item

    Связь товаров и промокодов

* id — id записи
* item_id — id товара
* promocode_id — id промокода

**Соотношение:**  
{id} → {item_id, promocode_id}
{item_id, promocode_id} → {id}
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓

## review

    Отзывы пользователей о магазинах

* id — id отзыва
* user_id — id пользователя, оставившего отзыв
* store_id — id магазина
* rating — оценка (0.0 – 5.0)
* comment — текст комментария

**Соотношение:**  
{id} → {user_id, store_id, rating, comment}   
1NF: ✓ 2NF: ✓ 3NF: ✓ BCNF: ✓   
