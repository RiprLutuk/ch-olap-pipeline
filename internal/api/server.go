// Package api exposes a small HTTP control plane for the generator.
//
// Endpoints:
//
//	GET  /healthz              liveness
//	GET  /metrics              Prometheus metrics
//	GET  /api/v1/stats         generator counters
//	POST /api/v1/generator/start
//	POST /api/v1/generator/stop
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RiprLutuk/ch-olap-pipeline/internal/db"
	"github.com/RiprLutuk/ch-olap-pipeline/internal/generator"
)

// Server holds the HTTP handlers.
type Server struct {
	svc   *generator.Service
	pools map[db.Target]*db.Pool
}

// NewServer builds a Server.
func NewServer(svc *generator.Service, pools map[db.Target]*db.Pool) *Server {
	return &Server{svc: svc, pools: pools}
}

// Routes returns an http.Handler with all routes mounted.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /metrics", s.metrics)
	mux.HandleFunc("GET /api/v1/stats", s.stats)
	mux.HandleFunc("POST /api/v1/generator/start", s.start)
	mux.HandleFunc("POST /api/v1/generator/stop", s.stop)
	return mux
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// metrics renders Prometheus exposition format from in-memory counters.
func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	var b []byte
	b = append(b, "# HELP generator_inserts_total Total rows inserted per target.\n"...)
	b = append(b, "# TYPE generator_inserts_total counter\n"...)
	for t, p := range s.pools {
		ins, _ := p.Stats()
		b = append(b, fmt.Sprintf("generator_inserts_total{target=%q} %d\n", string(t), ins)...)
	}
	b = append(b, "# HELP generator_updates_total Total rows updated per target.\n"...)
	b = append(b, "# TYPE generator_updates_total counter\n"...)
	for t, p := range s.pools {
		_, upd := p.Stats()
		b = append(b, fmt.Sprintf("generator_updates_total{target=%q} %d\n", string(t), upd)...)
	}
	b = append(b, "# HELP generator_running 1 if generator is running, 0 otherwise.\n"...)
	b = append(b, "# TYPE generator_running gauge\n"...)
	running := 0
	if s.svc.IsRunning() {
		running = 1
	}
	b = append(b, fmt.Sprintf("generator_running %d\n", running)...)
	_, _ = w.Write(b)
}

type statsResponse struct {
	Targets map[string]targetStats `json:"targets"`
}

type targetStats struct {
	Inserts uint64 `json:"inserts"`
	Updates uint64 `json:"updates"`
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	out := statsResponse{Targets: make(map[string]targetStats)}
	for t, p := range s.pools {
		ins, upd := p.Stats()
		out.Targets[string(t)] = targetStats{Inserts: ins, Updates: upd}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	s.svc.Start(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"started"}`))
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	s.svc.Stop()
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"stopped"}`))
}
