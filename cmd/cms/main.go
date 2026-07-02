// Command cms provides a lightweight web control panel and business CRUD for the demo pipeline.
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type app struct {
	generatorURL   string
	clickhouseURL  string
	clickhouseUser string
	clickhousePass string
	pgDB           *sqlx.DB
	mysqlDB        *sqlx.DB
	httpClient     *http.Client
}

type dashboardData struct {
	Now            string
	Tab            string
	Target         string
	GeneratorURL   string
	ClickhouseURL  string
	GeneratorStats string
	ClickhouseRows string
	Query          string
	QueryResult    string
	Customers      []customerRow
	Products       []productRow
	Orders         []orderRow
	Message        string
	Error          string
}

type customerRow struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	City  string `db:"city"`
}

type productRow struct {
	ID       int64   `db:"id"`
	Name     string  `db:"name"`
	Category string  `db:"category"`
	Price    float64 `db:"price"`
	Stock    int32   `db:"stock"`
}

type orderRow struct {
	ID          int64   `db:"id"`
	CustomerID  int64   `db:"customer_id"`
	Status      string  `db:"status"`
	TotalAmount float64 `db:"total_amount"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgDSN := getenv("CMS_PG_DSN", "postgres://shop:shop@postgres:5432/shop?sslmode=disable")
	mysqlDSN := getenv("CMS_MYSQL_DSN", "shop:shop@tcp(mysql:3306)/shop")

	a := &app{
		generatorURL:   getenv("CMS_GENERATOR_URL", "http://generator:8080"),
		clickhouseURL:  getenv("CMS_CLICKHOUSE_URL", "http://clickhouse:8123"),
		clickhouseUser: getenv("CMS_CLICKHOUSE_USER", "analytics"),
		clickhousePass: getenv("CMS_CLICKHOUSE_PASSWORD", "analytics"),
		httpClient:     &http.Client{Timeout: 8 * time.Second},
	}

	if db, err := openDB(ctx, "pgx", pgDSN); err == nil {
		a.pgDB = db
	} else {
		log.Printf("cms postgres warn: %v", err)
	}
	if db, err := openDB(ctx, "mysql", mysqlDSN); err == nil {
		a.mysqlDB = db
	} else {
		log.Printf("cms mysql warn: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.dashboard)
	mux.HandleFunc("POST /generator/start", a.generatorAction("/api/v1/generator/start"))
	mux.HandleFunc("POST /generator/stop", a.generatorAction("/api/v1/generator/stop"))
	mux.HandleFunc("POST /query", a.query)

	mux.HandleFunc("POST /customers/add", a.addCustomer)
	mux.HandleFunc("POST /products/add", a.addProduct)
	mux.HandleFunc("POST /orders/status", a.updateOrderStatus)

	addr := getenv("CMS_HTTP_ADDR", ":8080")
	log.Printf("cms listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func openDB(ctx context.Context, driver, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (a *app) dbForTarget(target string) *sqlx.DB {
	if strings.ToLower(target) == "mysql" {
		return a.mysqlDB
	}
	return a.pgDB
}

func (a *app) loadSummary(data *dashboardData) {
	stats, err := a.get(a.generatorURL + "/api/v1/stats")
	if err != nil {
		data.Error = appendErr(data.Error, fmt.Sprintf("generator stats: %v", err))
	} else {
		data.GeneratorStats = prettyJSON(stats)
	}
	rows, err := a.clickhouse("SELECT database, table, sum(rows) AS rows FROM system.parts WHERE active GROUP BY database, table ORDER BY database, table FORMAT PrettyCompact")
	if err != nil {
		data.Error = appendErr(data.Error, fmt.Sprintf("clickhouse rows: %v", err))
	} else {
		data.ClickhouseRows = rows
	}
}

func (a *app) dashboard(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "overview"
	}
	target := r.URL.Query().Get("target")
	if target == "" {
		target = "postgres"
	}

	data := dashboardData{
		Now:           time.Now().Format(time.RFC3339),
		Tab:           tab,
		Target:        target,
		GeneratorURL:  a.generatorURL,
		ClickhouseURL: a.clickhouseURL,
		Query:         defaultQuery,
	}

	a.loadSummary(&data)

	db := a.dbForTarget(target)
	if db != nil {
		_ = db.SelectContext(r.Context(), &data.Customers, "SELECT id, name, email, city FROM customers ORDER BY id DESC LIMIT 10")
		_ = db.SelectContext(r.Context(), &data.Products, "SELECT id, name, category, price, stock FROM products ORDER BY id DESC LIMIT 10")
		_ = db.SelectContext(r.Context(), &data.Orders, "SELECT id, customer_id, status, total_amount FROM orders ORDER BY id DESC LIMIT 10")
	}

	render(w, data)
}

func (a *app) generatorAction(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := a.post(a.generatorURL + path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		http.Redirect(w, r, "/?tab=overview", http.StatusSeeOther)
	}
}

func (a *app) query(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.FormValue("sql"))
	data := dashboardData{
		Now:           time.Now().Format(time.RFC3339),
		Tab:           "analytics",
		Target:        r.FormValue("target"),
		GeneratorURL:  a.generatorURL,
		ClickhouseURL: a.clickhouseURL,
		Query:         q,
	}
	if data.Target == "" {
		data.Target = "postgres"
	}
	if q == "" {
		data.Error = "query is empty"
		a.loadSummary(&data)
		render(w, data)
		return
	}
	out, err := a.clickhouse(q)
	if err != nil {
		data.Error = err.Error()
	} else {
		data.QueryResult = out
	}
	a.loadSummary(&data)
	render(w, data)
}

func (a *app) addCustomer(w http.ResponseWriter, r *http.Request) {
	target := r.FormValue("target")
	db := a.dbForTarget(target)
	if db == nil {
		http.Redirect(w, r, "/?tab=customers&target="+target, http.StatusSeeOther)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	city := strings.TrimSpace(r.FormValue("city"))

	var err error
	if strings.ToLower(target) == "mysql" {
		_, err = db.Exec("INSERT INTO customers (name, email, city) VALUES (?, ?, ?)", name, email, city)
	} else {
		_, err = db.Exec("INSERT INTO customers (name, email, city) VALUES ($1, $2, $3)", name, email, city)
	}
	if err != nil {
		log.Printf("addCustomer err: %v", err)
	}
	http.Redirect(w, r, "/?tab=customers&target="+target, http.StatusSeeOther)
}

func (a *app) addProduct(w http.ResponseWriter, r *http.Request) {
	target := r.FormValue("target")
	db := a.dbForTarget(target)
	if db == nil {
		http.Redirect(w, r, "/?tab=products&target="+target, http.StatusSeeOther)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	category := strings.TrimSpace(r.FormValue("category"))
	price, _ := strconv.ParseFloat(strings.TrimSpace(r.FormValue("price")), 64)
	stock, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("stock")), 10, 32)

	var err error
	if strings.ToLower(target) == "mysql" {
		_, err = db.Exec("INSERT INTO products (name, category, price, stock) VALUES (?, ?, ?, ?)", name, category, price, int32(stock))
	} else {
		_, err = db.Exec("INSERT INTO products (name, category, price, stock) VALUES ($1, $2, $3, $4)", name, category, price, int32(stock))
	}
	if err != nil {
		log.Printf("addProduct err: %v", err)
	}
	http.Redirect(w, r, "/?tab=products&target="+target, http.StatusSeeOther)
}

func (a *app) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	target := r.FormValue("target")
	db := a.dbForTarget(target)
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	status := strings.TrimSpace(r.FormValue("status"))

	if db != nil && id > 0 && status != "" {
		var err error
		if strings.ToLower(target) == "mysql" {
			_, err = db.Exec("UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?", status, id)
		} else {
			_, err = db.Exec("UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
		}
		if err != nil && err != sql.ErrNoRows {
			log.Printf("updateOrderStatus err: %v", err)
		}
	}
	http.Redirect(w, r, "/?tab=orders&target="+target, http.StatusSeeOther)
}

func (a *app) get(u string) ([]byte, error) {
	resp, err := a.httpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	return body, nil
}

func (a *app) post(u string) ([]byte, error) {
	resp, err := a.httpClient.Post(u, "application/json", bytes.NewReader(nil))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	return body, nil
}

func (a *app) clickhouse(sqlStr string) (string, error) {
	u := a.clickhouseURL + "/?user=" + url.QueryEscape(a.clickhouseUser) + "&password=" + url.QueryEscape(a.clickhousePass)
	resp, err := a.httpClient.Post(u, "text/plain", strings.NewReader(sqlStr))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	return string(body), nil
}

func render(w http.ResponseWriter, data dashboardData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Execute(w, data); err != nil {
		log.Printf("render: %v", err)
	}
}

func prettyJSON(b []byte) string {
	var dst bytes.Buffer
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		return string(b)
	}
	return dst.String()
}

func appendErr(existing, msg string) string {
	if existing == "" {
		return msg
	}
	return existing + "\n" + msg
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

const defaultQuery = `SELECT
  database,
  table,
  sum(rows) AS rows
FROM system.parts
WHERE active
GROUP BY database, table
ORDER BY database, table
FORMAT PrettyCompact`

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CH OLAP Pipeline CMS</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, sans-serif; }
    body { margin: 0; background: #08111f; color: #e5eefc; }
    header { padding: 20px 32px; background: linear-gradient(135deg, #0f172a, #111827); border-bottom: 1px solid #1f2937; display: flex; justify-content: space-between; align-items: center; }
    h1 { margin: 0; font-size: 24px; }
    nav { display: flex; gap: 8px; margin-top: 12px; }
    nav a { padding: 8px 14px; border-radius: 8px; background: #1e293b; color: #cbd5e1; text-decoration: none; font-weight: 600; font-size: 14px; }
    nav a.active { background: #2563eb; color: #fff; }
    main { padding: 24px 32px; display: grid; gap: 20px; }
    .grid { display: grid; gap: 20px; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); }
    .card { background: #0f172a; border: 1px solid #243047; border-radius: 16px; padding: 18px; box-shadow: 0 10px 30px rgba(0,0,0,.18); }
    .muted { color: #94a3b8; font-size: 13px; }
    button, .linkbtn { background: #2563eb; color: white; border: 0; padding: 8px 12px; border-radius: 8px; cursor: pointer; font-weight: 700; text-decoration: none; display: inline-block; }
    button.stop { background: #dc2626; }
    textarea { width: 100%; min-height: 140px; background: #020617; color: #dbeafe; border: 1px solid #334155; border-radius: 12px; padding: 12px; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
    pre { overflow: auto; background: #020617; border: 1px solid #1e293b; border-radius: 12px; padding: 14px; color: #bfdbfe; min-height: 80px; }
    .error { background: #450a0a; border-color: #991b1b; color: #fecaca; white-space: pre-wrap; }
    form.inline { display: inline; margin-right: 8px; }
    table { width: 100%; border-collapse: collapse; margin-top: 12px; }
    th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid #1e293b; font-size: 14px; }
    th { color: #94a3b8; }
    input, select { background: #020617; border: 1px solid #334155; color: #fff; padding: 8px 10px; border-radius: 8px; margin-right: 8px; }
    .toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
  </style>
</head>
<body>
<header>
  <div>
    <h1>CH OLAP Pipeline CMS</h1>
    <nav>
      <a href="/?tab=overview&target={{.Target}}" class="{{if eq .Tab "overview"}}active{{end}}">Overview</a>
      <a href="/?tab=customers&target={{.Target}}" class="{{if eq .Tab "customers"}}active{{end}}">Customers</a>
      <a href="/?tab=products&target={{.Target}}" class="{{if eq .Tab "products"}}active{{end}}">Products</a>
      <a href="/?tab=orders&target={{.Target}}" class="{{if eq .Tab "orders"}}active{{end}}">Orders</a>
      <a href="/?tab=analytics&target={{.Target}}" class="{{if eq .Tab "analytics"}}active{{end}}">Analytics</a>
    </nav>
  </div>
  <div>
    <span class="muted">Target DB: </span>
    <a href="/?tab={{.Tab}}&target=postgres" class="linkbtn {{if ne .Target "mysql"}}stop{{end}}">Postgres</a>
    <a href="/?tab={{.Tab}}&target=mysql" class="linkbtn {{if eq .Target "mysql"}}stop{{end}}">MySQL</a>
  </div>
</header>
<main>
  {{if .Error}}<pre class="card error">{{.Error}}</pre>{{end}}

  {{if eq .Tab "overview"}}
  <section class="grid">
    <div class="card">
      <h2>Generator Control</h2>
      <p class="muted">{{.GeneratorURL}}</p>
      <form class="inline" method="post" action="/generator/start"><button>Start Generator</button></form>
      <form class="inline" method="post" action="/generator/stop"><button class="stop">Stop Generator</button></form>
      <h3>Stats</h3>
      <pre>{{.GeneratorStats}}</pre>
    </div>
    <div class="card">
      <h2>ClickHouse Row Counts</h2>
      <p class="muted">{{.ClickhouseURL}}</p>
      <pre>{{.ClickhouseRows}}</pre>
    </div>
  </section>
  {{end}}

  {{if eq .Tab "customers"}}
  <section class="card">
    <div class="toolbar">
      <h2>Customers ({{.Target}})</h2>
      <form method="post" action="/customers/add" class="inline">
        <input type="hidden" name="target" value="{{.Target}}">
        <input type="text" name="name" placeholder="Name" required>
        <input type="email" name="email" placeholder="Email" required>
        <input type="text" name="city" placeholder="City" required>
        <button>Add Customer</button>
      </form>
    </div>
    <table>
      <thead><tr><th>ID</th><th>Name</th><th>Email</th><th>City</th></tr></thead>
      <tbody>
      {{range .Customers}}
        <tr><td>{{.ID}}</td><td>{{.Name}}</td><td>{{.Email}}</td><td>{{.City}}</td></tr>
      {{else}}
        <tr><td colspan="4" class="muted">No customers found</td></tr>
      {{end}}
      </tbody>
    </table>
  </section>
  {{end}}

  {{if eq .Tab "products"}}
  <section class="card">
    <div class="toolbar">
      <h2>Products ({{.Target}})</h2>
      <form method="post" action="/products/add" class="inline">
        <input type="hidden" name="target" value="{{.Target}}">
        <input type="text" name="name" placeholder="Product Name" required>
        <input type="text" name="category" placeholder="Category" required>
        <input type="number" step="0.01" name="price" placeholder="Price" required>
        <input type="number" name="stock" placeholder="Stock" required>
        <button>Add Product</button>
      </form>
    </div>
    <table>
      <thead><tr><th>ID</th><th>Name</th><th>Category</th><th>Price</th><th>Stock</th></tr></thead>
      <tbody>
      {{range .Products}}
        <tr><td>{{.ID}}</td><td>{{.Name}}</td><td>{{.Category}}</td><td>${{.Price}}</td><td>{{.Stock}}</td></tr>
      {{else}}
        <tr><td colspan="5" class="muted">No products found</td></tr>
      {{end}}
      </tbody>
    </table>
  </section>
  {{end}}

  {{if eq .Tab "orders"}}
  <section class="card">
    <h2>Recent Orders ({{.Target}})</h2>
    <table>
      <thead><tr><th>ID</th><th>Customer ID</th><th>Status</th><th>Total Amount</th><th>Action</th></tr></thead>
      <tbody>
      {{range .Orders}}
        <tr>
          <td>{{.ID}}</td><td>{{.CustomerID}}</td><td>{{.Status}}</td><td>${{.TotalAmount}}</td>
          <td>
            <form method="post" action="/orders/status" class="inline">
              <input type="hidden" name="target" value="{{$.Target}}">
              <input type="hidden" name="id" value="{{.ID}}">
              <select name="status">
                <option value="PLACED" {{if eq .Status "PLACED"}}selected{{end}}>PLACED</option>
                <option value="PAID" {{if eq .Status "PAID"}}selected{{end}}>PAID</option>
                <option value="SHIPPED" {{if eq .Status "SHIPPED"}}selected{{end}}>SHIPPED</option>
                <option value="DELIVERED" {{if eq .Status "DELIVERED"}}selected{{end}}>DELIVERED</option>
                <option value="CANCELLED" {{if eq .Status "CANCELLED"}}selected{{end}}>CANCELLED</option>
              </select>
              <button type="submit">Update</button>
            </form>
          </td>
        </tr>
      {{else}}
        <tr><td colspan="5" class="muted">No orders found</td></tr>
      {{end}}
      </tbody>
    </table>
  </section>
  {{end}}

  {{if eq .Tab "analytics"}}
  <section class="card">
    <h2>Ad-hoc ClickHouse SQL</h2>
    <form method="post" action="/query">
      <input type="hidden" name="target" value="{{.Target}}">
      <textarea name="sql">{{.Query}}</textarea>
      <p><button>Run Query</button></p>
    </form>
    <pre>{{.QueryResult}}</pre>
  </section>
  {{end}}
</main>
</body>
</html>`))
