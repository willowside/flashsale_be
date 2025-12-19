package domain

type Product struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	Price     int    `db:"price"`
	Stock     int    `db:"stock"`
	CreatedAt string `db:"created_at"`
}
