select acc.name, r.rating, r.comment, r.created_at
from review r
         left join account acc on r.user_id = acc.id
where r.store_id = $1
order by r.created_at desc