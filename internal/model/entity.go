// Package model holds shared entity definitions.
package model

// Customer represents a customer row in OLTP.
type Customer struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	City  string `db:"city"`
}

// Product represents a product row in OLTP.
type Product struct {
	ID       int64   `db:"id"`
	Name     string  `db:"name"`
	Category string  `db:"category"`
	Price    float64 `db:"price"`
	Stock    int32   `db:"stock"`
}

// Order represents an order row in OLTP.
type Order struct {
	ID          int64   `db:"id"`
	CustomerID  int64   `db:"customer_id"`
	Status      string  `db:"status"`
	TotalAmount float64 `db:"total_amount"`
}
