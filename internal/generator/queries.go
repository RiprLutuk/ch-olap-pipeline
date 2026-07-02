package generator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/RiprLutuk/ch-olap-pipeline/internal/db"
	"github.com/RiprLutuk/ch-olap-pipeline/internal/model"
)

var cities = []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Yogyakarta", "Bali", "Makassar"}
var firstNames = []string{"Andi", "Budi", "Citra", "Dewi", "Eko", "Fitri", "Galih", "Hana", "Indra", "Jaka"}
var lastNames = []string{"Wijaya", "Pratama", "Saputra", "Lestari", "Nugroho", "Sukma", "Maulana", "Hakim"}

var statuses = []string{"PLACED", "PAID", "SHIPPED", "DELIVERED", "CANCELLED"}

// State machine transition probabilities. Aligned with the reference repo.
var transitions = map[string][]struct {
	Next string
	P    float64
}{
	"PLACED":  {{"PAID", 0.90}, {"CANCELLED", 0.10}},
	"PAID":    {{"SHIPPED", 0.95}, {"CANCELLED", 0.05}},
	"SHIPPED": {{"DELIVERED", 0.99}, {"CANCELLED", 0.01}},
}

func (s *Service) createCustomer(ctx context.Context, pool *db.Pool, target db.Target) error {
	c := model.Customer{
		Name:  fmt.Sprintf("%s %s", pick(firstNames), pick(lastNames)),
		Email: fmt.Sprintf("user%d_%s@example.test", rand.Int63(), pick(cities)),
		City:  pick(cities),
	}
	q := insertCustomerQuery(target)
	_, err := pool.DB.ExecContext(ctx, q, c.Name, c.Email, c.City)
	if err != nil {
		return err
	}
	pool.IncInsert(1)
	return nil
}

func (s *Service) createOrder(ctx context.Context, pool *db.Pool, target db.Target) error {
	customerID, err := pickRandomCustomerID(ctx, pool, target)
	if err != nil {
		return err
	}
	if customerID == 0 {
		// No customers yet — bootstrap one.
		if err := s.createCustomer(ctx, pool, target); err != nil {
			return err
		}
		customerID, err = pickRandomCustomerID(ctx, pool, target)
		if err != nil || customerID == 0 {
			return err
		}
	}

	productID, price, err := pickRandomProduct(ctx, pool, target)
	if err != nil {
		return err
	}
	qty := rand.Intn(3) + 1
	total := price * float64(qty)

	insertOrder := insertOrderQuery(target)
	var orderID int64
	if target == db.TargetPostgres {
		row := pool.DB.QueryRowxContext(ctx, insertOrder+" RETURNING id", customerID, "PLACED", total)
		if err := row.Scan(&orderID); err != nil {
			return err
		}
	} else {
		res, err := pool.DB.ExecContext(ctx, insertOrder, customerID, "PLACED", total)
		if err != nil {
			return err
		}
		orderID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}

	insertItem := insertOrderItemQuery(target)
	_, err = pool.DB.ExecContext(ctx, insertItem, orderID, productID, qty, price)
	if err != nil {
		return err
	}
	pool.IncInsert(2)
	return nil
}

func (s *Service) advanceOrders(ctx context.Context, pool *db.Pool, target db.Target) error {
	drains := map[string]float64{
		"PLACED":  0.08,
		"PAID":    0.20,
		"SHIPPED": 0.30,
	}
	for status, drain := range drains {
		var count int
		if err := pool.DB.GetContext(ctx, &count, countByStatusQuery(target), status); err != nil {
			return err
		}
		if count == 0 {
			continue
		}
		n := int(float64(count) * drain)
		if n < 1 {
			n = 1
		}
		rows, err := pool.DB.QueryxContext(ctx, pickByStatusQuery(target), status, n)
		if err != nil {
			return err
		}
		var ids []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			ids = append(ids, id)
		}
		rows.Close()
		for _, id := range ids {
			next := nextStatus(status)
			if next == "" {
				continue
			}
			if _, err := pool.DB.ExecContext(ctx, updateStatusQuery(target), next, id); err != nil {
				return err
			}
			pool.IncUpdate(1)
		}
	}
	return nil
}

func nextStatus(current string) string {
	opts, ok := transitions[current]
	if !ok {
		return ""
	}
	r := rand.Float64()
	cum := 0.0
	for _, o := range opts {
		cum += o.P
		if r < cum {
			return o.Next
		}
	}
	return ""
}

func pickRandomCustomerID(ctx context.Context, pool *db.Pool, target db.Target) (int64, error) {
	var id int64
	err := pool.DB.GetContext(ctx, &id, pickRandomCustomerQuery(target))
	return id, err
}

func pickRandomProduct(ctx context.Context, pool *db.Pool, target db.Target) (int64, float64, error) {
	var id int64
	var price float64
	err := pool.DB.QueryRowxContext(ctx, pickRandomProductQuery(target)).Scan(&id, &price)
	return id, price, err
}

func pick[T any](s []T) T { return s[rand.Intn(len(s))] }

// --- Per-target SQL builders -----------------------------------------

func paramStyle(target db.Target, n int) string {
	if target == db.TargetPostgres {
		return strings.Repeat("$", 0) + strings.Repeat(",$", n)[1:]
	}
	return "?"
}

func insertCustomerQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "INSERT INTO customers (name, email, city) VALUES ($1, $2, $3)"
	}
	return "INSERT INTO customers (name, email, city) VALUES (?, ?, ?)"
}

func insertOrderQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "INSERT INTO orders (customer_id, status, total_amount) VALUES ($1, $2, $3)"
	}
	return "INSERT INTO orders (customer_id, status, total_amount) VALUES (?, ?, ?)"
}

func insertOrderItemQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, $4)"
	}
	return "INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)"
}

func pickRandomCustomerQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "SELECT id FROM customers ORDER BY RANDOM() LIMIT 1"
	}
	return "SELECT id FROM customers ORDER BY RAND() LIMIT 1"
}

func pickRandomProductQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "SELECT id, price FROM products ORDER BY RANDOM() LIMIT 1"
	}
	return "SELECT id, price FROM products ORDER BY RAND() LIMIT 1"
}

func countByStatusQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "SELECT COUNT(*) FROM orders WHERE status = $1"
	}
	return "SELECT COUNT(*) FROM orders WHERE status = ?"
}

func pickByStatusQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "SELECT id FROM orders WHERE status = $1 ORDER BY RANDOM() LIMIT $2"
	}
	return "SELECT id FROM orders WHERE status = ? ORDER BY RAND() LIMIT ?"
}

func updateStatusQuery(t db.Target) string {
	if t == db.TargetPostgres {
		return "UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2"
	}
	return "UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?"
}
