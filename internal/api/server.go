// Package api exposes a small HTTP control plane for the generator.
//
// Endpoints:
//
//	GET  /healthz              liveness
//	GET  /api/v1/stats         generator counters
//	POST /api/v1/generator/start
//	POST /api/v1/generator/stop
package api

import (
	"encoding/json"
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
	mux.HandleFunc("GET /api/v1/stats", s.stats)
	mux.HandleFunc("POST /api/v1/generator/start", s.start)
	mux.HandleFunc("POST /api/v1/generator/stop", s.stop)
	return mux
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
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
