select s.id,
       s.name,
       s.description,
       s.city_id,
       s.address,
       s.card_img,
       s.rating,
       s.open_at,
       s.closed_at,
       st.tag_id
from store s
         left join store_tag st on st.store_id = s.id
where s.id = $1  -- ← ИСПРАВЛЕНО: добавить s.