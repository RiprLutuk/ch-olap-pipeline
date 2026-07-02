// Package generator drives e-commerce load against OLTP sources.
//
// One worker goroutine per (target, lane) pair. Each tick:
//   - with low probability, create a new customer
//   - create N new orders
//   - advance existing orders through the status pipeline proportionally
//
// The advance model uses a per-status drain rate so the queue reaches
// equilibrium, mirroring a real e-commerce funnel.
package generator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/RiprLutuk/ch-olap-pipeline/internal/db"
)

// Config controls generator behavior.
type Config struct {
	TickInterval time.Duration
	Workers      int
	OrderPerTick int
	Targets      []db.Target
}

// Service is the long-running generator.
type Service struct {
	cfg    Config
	pools  map[db.Target]*db.Pool
	mu     sync.RWMutex
	stopCh chan struct{}
}

// NewService wires the generator against the supplied pools.
func NewService(cfg Config, pools map[db.Target]*db.Pool) *Service {
	return &Service{
		cfg:    cfg,
		pools:  pools,
		stopCh: make(chan struct{}),
	}
}

// Start launches the worker goroutines. Idempotent: subsequent calls are no-op.
func (s *Service) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.cfg.Targets {
		pool, ok := s.pools[t]
		if !ok {
			log.Printf("generator: target %s requested but no pool configured, skipping", t)
			continue
		}
		for w := 0; w < s.cfg.Workers; w++ {
			workerID := w
			target := t
			go s.runWorker(ctx, pool, target, workerID)
		}
	}
}

// Stop signals all workers to exit.
func (s *Service) Stop() {
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
}

func (s *Service) runWorker(ctx context.Context, pool *db.Pool, target db.Target, id int) {
	t := time.NewTicker(s.cfg.TickInterval)
	defer t.Stop()
	log.Printf("generator: worker %d/%s started", id, target)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-t.C:
			if err := s.tick(ctx, pool, target); err != nil {
				log.Printf("generator: tick %s#%d failed: %v", target, id, err)
			}
		}
	}
}

func (s *Service) tick(ctx context.Context, pool *db.Pool, target db.Target) error {
	// 1) occasionally create a customer
	if rand.Float64() < 0.30 {
		if err := s.createCustomer(ctx, pool, target); err != nil {
			return fmt.Errorf("create customer: %w", err)
		}
	}

	// 2) create N orders per tick
	for i := 0; i < s.cfg.OrderPerTick; i++ {
		if err := s.createOrder(ctx, pool, target); err != nil {
			return fmt.Errorf("create order: %w", err)
		}
	}

	// 3) advance order status proportionally
	if err := s.advanceOrders(ctx, pool, target); err != nil {
		return fmt.Errorf("advance orders: %w", err)
	}
	return nil
}
