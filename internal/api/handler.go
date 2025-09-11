package api

import (
	// "context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"yourmodule/internal/cache"
	"yourmodule/internal/db"
)

type Server struct {
	store *db.Store
	cache *cache.Cache
}

func NewServer(store *db.Store, cache *cache.Cache) *Server {
	return &Server{store: store, cache: cache}
}

func (s *Server) Routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/order/{order_uid}", s.GetOrder).Methods("GET")

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

	// 1) try cache
	if ord, ok := s.cache.Get(id); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ord)
		return
	}

	// 2) fallback to DB
	ctx := r.Context()
	ord, _, err := s.store.GetOrder(ctx, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// put into cache for future
	s.cache.Set(id, ord)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ord)
}
