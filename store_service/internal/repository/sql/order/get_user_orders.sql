select o.id          as id,
       o.status      as status,
       o.total_price as total,
       o.created_at  as created_at
from "order" o
where o.user_id = $1
order by o.created_at;