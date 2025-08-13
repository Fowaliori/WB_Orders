package cache

import (
	"l0/internal/models"
	"sync"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]models.Order
}

func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]models.Order),
	}
}

func (c *Cache) Set(uid string, order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[uid] = order
}

func (c *Cache) Get(uid string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.orders[uid]
	return order, exists
}

// GetAll возвращает все заказы в кэше (для проверки)
func (c *Cache) GetAll() map[string]models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Создаем копию map
	result := make(map[string]models.Order, len(c.orders))
	for k, v := range c.orders {
		result[k] = v
	}
	return result
}
