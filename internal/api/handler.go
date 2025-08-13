package api

import (
	"encoding/json"
	"log"
	"net/http"

	"l0/internal/cache"
	"l0/internal/db"
)

type Handler struct {
	cache *cache.Cache
	db    *db.Postgres
}

func NewHandler(cache *cache.Cache, db *db.Postgres) http.Handler {
	return &Handler{
		cache: cache,
		db:    db,
	}
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

	// 1. Сначала проверяем кэш
	if order, exists := h.cache.Get(uid); exists {
		if err := json.NewEncoder(w).Encode(order); err != nil {
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	// 2. Если нет в кэше, проверяем БД
	order, err := h.db.GetOrder(r.Context(), uid)
	if err != nil {
		log.Printf("Order %s not found in DB: %v", uid, err)
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	}

	// 3. Добавляем найденный заказ в кэш
	h.cache.Set(uid, *order)

	// 4. Возвращаем результат
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}
