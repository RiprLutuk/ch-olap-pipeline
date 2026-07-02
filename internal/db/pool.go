// Package db manages a multi-DB connection pool.
//
// One pool per target (Postgres, MySQL, future). Each pool exposes
// acquire/release semantics through sqlx and a `Ping` for healthcheck.
package db

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Target identifies a logical OLTP source.
type Target string

const (
	TargetPostgres Target = "postgres"
	TargetMySQL    Target = "mysql"
)

// Pool is a thin wrapper over sqlx.DB with metrics counters.
type Pool struct {
	Target Target
	DB     *sqlx.DB

	insertCount atomic.Uint64
	updateCount atomic.Uint64
}

// New opens a connection pool for the given target.
func New(ctx context.Context, target Target, dsn string) (*Pool, error) {
	driver := driverFor(target)
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", target, err)
	}
	// Sensible defaults; tune via env later.
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping %s: %w", target, err)
	}
	return &Pool{Target: target, DB: db}, nil
}

func driverFor(t Target) string {
	switch t {
	case TargetPostgres:
		return "pgx"
	case TargetMySQL:
		return "mysql"
	}
	return ""
}

// IncInsert bumps the insert counter.
func (p *Pool) IncInsert(n uint64) { p.insertCount.Add(n) }

// IncUpdate bumps the update counter.
func (p *Pool) IncUpdate(n uint64) { p.updateCount.Add(n) }

// Stats returns current counters for observability.
func (p *Pool) Stats() (inserts, updates uint64) {
	return p.insertCount.Load(), p.updateCount.Load()
}

// Close releases the pool.
func (p *Pool) Close() error { return p.DB.Close() }
