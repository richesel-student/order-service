package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"yourmodule/internal/models"

	"github.com/gorilla/mux"
)

/************* INTERFACES *************/

type Store interface {
	GetOrder(ctx context.Context, id string) (models.Order, []byte, error)
}

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, val interface{}, ttl time.Duration)
}

/************* SERVER *************/

type Server struct {
	store Store
	cache Cache

	slowMu sync.RWMutex
}

func NewServer(store Store, cache Cache) *Server {
	return &Server{store: store, cache: cache}
}

func (s *Server) Routes() http.Handler {
	r := mux.NewRouter()

	// pprof handlers
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	r.HandleFunc("/order/{order_uid}", s.GetOrder).Methods(http.MethodGet)

	// catch-all route MUST be last
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	return r
}

func (s *Server) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["order_uid"]

	if id == "" {
		http.Error(w, "order_uid required", http.StatusBadRequest)
		return
	}

	// --- optimized bottleneck section ---
	s.slowMu.RLock()
	defer s.slowMu.RUnlock()

	// reduced string operations
	slowStr := "order_" + id
	for i := 0; i < 50; i++ {
		slowStr = strings.ToUpper(slowStr)
	}
	_ = slowStr

	// reduced allocations
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteString(id)

	// reduced goroutine fanout
	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			for j := 0; j < 1000; j++ {
				_ = strconv.Itoa(i)
				_ = strconv.Itoa(j)
			}
		}(i)
	}

	wg.Wait()
	// --- optimized bottleneck section end ---

	// cache
	if v, ok := s.cache.Get(id); ok {
		if ord, ok := v.(models.Order); ok {
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(ord); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			return
		}
	}

	// db
	ord, _, err := s.store.GetOrder(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	s.cache.Set(id, ord, time.Minute)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(ord); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}