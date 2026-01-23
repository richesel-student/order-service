package api

import (
	"context"
	"encoding/json"
	"net/http"
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
}

func NewServer(store Store, cache Cache) *Server {
	return &Server{store: store, cache: cache}
}

func (s *Server) Routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/order/{order_uid}", s.GetOrder).Methods(http.MethodGet)
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

	if v, ok := s.cache.Get(id); ok {
		if ord, ok := v.(models.Order); ok {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(ord); err != nil {
				// обработка ошибки: отправляем 500 и логируем
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	// 2) db
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
