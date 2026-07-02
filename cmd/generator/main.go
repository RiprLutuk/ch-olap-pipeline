// Command generator is the entry point for the multi-DB load generator.
//
// Env vars:
//
//	GEN_HTTP_ADDR   listen address (default ":8080")
//	GEN_PG_DSN      postgres DSN
//	GEN_MYSQL_DSN   mysql DSN
//	GEN_TICK_MS     tick interval in ms (default 1000)
//	GEN_WORKERS     worker goroutines per target (default 4)
//	GEN_ORDER_TICK  orders per tick (default 5)
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/RiprLutuk/ch-olap-pipeline/internal/api"
	"github.com/RiprLutuk/ch-olap-pipeline/internal/db"
	"github.com/RiprLutuk/ch-olap-pipeline/internal/generator"
)

func main() {
	addr := getEnv("GEN_HTTP_ADDR", ":8080")
	tickMs, _ := strconv.Atoi(getEnv("GEN_TICK_MS", "1000"))
	workers, _ := strconv.Atoi(getEnv("GEN_WORKERS", "4"))
	orderTick, _ := strconv.Atoi(getEnv("GEN_ORDER_TICK", "5"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pools := make(map[db.Target]*db.Pool)
	var targets []db.Target

	if dsn := os.Getenv("GEN_PG_DSN"); dsn != "" {
		p, err := db.New(ctx, db.TargetPostgres, dsn)
		if err != nil {
			log.Fatalf("postgres pool: %v", err)
		}
		pools[db.TargetPostgres] = p
		targets = append(targets, db.TargetPostgres)
		defer p.Close()
		log.Printf("postgres pool ready")
	}
	if dsn := os.Getenv("GEN_MYSQL_DSN"); dsn != "" {
		p, err := db.New(ctx, db.TargetMySQL, dsn)
		if err != nil {
			log.Fatalf("mysql pool: %v", err)
		}
		pools[db.TargetMySQL] = p
		targets = append(targets, db.TargetMySQL)
		defer p.Close()
		log.Printf("mysql pool ready")
	}
	if len(pools) == 0 {
		log.Fatalf("no DSN provided: set GEN_PG_DSN and/or GEN_MYSQL_DSN")
	}

	svc := generator.NewService(generator.Config{
		TickInterval: time.Duration(tickMs) * time.Millisecond,
		Workers:      workers,
		OrderPerTick: orderTick,
		Targets:      targets,
	}, pools)

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewServer(svc, pools).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("generator HTTP listening on %s (workers=%d, tick=%dms, targets=%v)",
			addr, workers, tickMs, targets)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutdown signal received")
	svc.Stop()
	shutdownCtx, sc := context.WithTimeout(context.Background(), 5*time.Second)
	defer sc()
	_ = srv.Shutdown(shutdownCtx)
}

func getEnv(k, fallback string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return fallback
}
