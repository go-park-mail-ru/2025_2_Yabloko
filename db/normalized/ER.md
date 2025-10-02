```mermaid
erDiagram
    account {
        uuid id PK
        text email 
        text password 
        text name 
        text phone 
        uuid city_id FK
        text address 
    }

    city {
        uuid id PK
        text name
    }

    store {
        uuid id PK
        text name
        text description
        uuid city_id FK
        text address
        text card_img
        numeric rating
        timetz open_at
        timetz closed_at
    }

    friend {
        uuid account_1_id FK
        uuid account_2_id FK
    }

    cart {
        uuid id PK
        uuid user_id FK
    }

    cart_item {
        uuid id PK
        uuid cart_id FK
        uuid store_item_id FK
        integer quantity
    }
    
    order {
        uuid id PK
        uuid user_id FK
        uuid store_id FK
        numeric total_price
        text status
    }
    
    order_item {
        uuid id PK
        uuid order_id FK
        uuid store_item_id FK
        numeric price
        integer quantity
    }
    
    tag {
        uuid id PK
        text name
    }
    
    store_tag {
        uuid store_id FK
        uuid tag_id FK
    }
    
    category {
        uuid id PK
        text name
    }
    
    store_category {
        uuid store_id FK
        uuid category_id FK
    }
    
    item {
        uuid id PK
        text name
        text description
        text card_img
    }
    
    store_item {
        uuid id PK
        uuid store_id FK
        uuid item_id FK
        numeric price
    }
    
    promotion {
        uuid id PK
        text name
        text type
        numeric value
        text code
        timestamptz start_time
        timestamptz end_time
    }
    
    promotion_item {
        uuid item_id FK
        uuid promotion_id FK
    }
    
    type {
        uuid id PK
        text name
    }
    
    item_type {
        uuid item_id FK
        uuid type_id FK
    }

    city ||--o{ account : city_id
    city ||--o{ store : city_id

    account ||--o{ friend : account_1_id
    account ||--o{ friend : account_2_id
    account ||--o{ cart : user_id
    account ||--o{ order : user_id

    cart ||--o{ cart_item : cart_id
    store_item ||--o{ cart_item : store_item_id

    store ||--o{ order : store_id
    order ||--o{ order_item : order_id
    store_item ||--o{ order_item : store_item_id

    store ||--o{ store_item : store_id
    store ||--o{ store_tag : store_id
    store ||--o{ store_category : store_id

    item ||--o{ store_item : item_id
    item ||--o{ promotion_item : item_id
    item ||--o{ item_type : item_id

    tag ||--o{ store_tag : tag_id
    category ||--o{ store_category : category_id
    promotion ||--o{ promotion_item : promotion_id
    type ||--o{ item_type : type_id
```