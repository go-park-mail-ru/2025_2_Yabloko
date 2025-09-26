package auth

const (
	addUser = `
		insert into account (id, email, hash)
		values ($1, $2, $3);
	`

	getUser = `
		select hash from account
		where email = $1;
	`

	deleteUser = `
		delete from account
		where email = $1;
	`
)
