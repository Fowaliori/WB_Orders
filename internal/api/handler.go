package api

import (
	"encoding/json"
	"net/http"

	"l0/internal/cache"
)

type Handler struct {
	cache *cache.Cache
}

func NewHandler(cache *cache.Cache) http.Handler {
	return &Handler{cache: cache}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/order":
		h.getOrder(w, r)
	default:
		http.FileServer(http.Dir("./web")).ServeHTTP(w, r)
	}
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, `{"error": "Order UID is required"}`, http.StatusBadRequest)
		return
	}

	order, exists := h.cache.Get(uid)
	if !exists {
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}
