package store

const (
	addStore = `
		insert into store (id, name, description, city_id, address, card_img, rating, open_at, closed_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	//getStore = `
	//	select store.id, store.name, description, city_id.name, address, card_img, rating, open_at, closed_at from store
	//	left join city on city.id = store.city_id
	//	where store.id > $2
	//	order by store.id
	//	limit $1;
	//`
	getStore = `
		select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store
		where id > $2
		order by id
		limit $1;
	`

	//getStoreFirst = `
	//	select store.id, store.name, description, city.name, address, card_img, rating, open_at, closed_at from store
	//	left join city on city.id = store.city_id
	//	order by store.id
	//	limit $1;
	//`
	getStoreFirst = `
		select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store
		order by id
		limit $1;
	`

	updateUser = `
	//TODO
	`
)
