package cache

import (
	"testing"
	"time"

	"l0/internal/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	memoryCache := NewCache().(*ShardedCache)

	order := models.Order{
		OrderUID:    "test-123",
		TrackNumber: "WBIL123",
		DateCreated: time.Now(),
	}

	memoryCache.Set(order.OrderUID, order)
	retrieved, exists := memoryCache.Get(order.OrderUID)

	assert.True(t, exists)
	assert.Equal(t, order.OrderUID, retrieved.OrderUID)
	assert.Equal(t, order.TrackNumber, retrieved.TrackNumber)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheService := NewCache().(*ShardedCache)

	_, exists := cacheService.Get("non-existent")
	assert.False(t, exists)
}

func TestMemoryCache_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheService := NewCache().(*ShardedCache)

	order1 := models.Order{OrderUID: "test-1", TrackNumber: "WBIL1"}
	order2 := models.Order{OrderUID: "test-2", TrackNumber: "WBIL2"}

	cacheService.Set(order1.OrderUID, order1)
	cacheService.Set(order2.OrderUID, order2)

	allOrders := cacheService.GetAll()
	assert.Len(t, allOrders, 2)
	assert.Contains(t, allOrders, "test-1")
	assert.Contains(t, allOrders, "test-2")
}

func TestMemoryCache_Remove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheService := NewCache().(*ShardedCache)

	order := models.Order{OrderUID: "test-remove", TrackNumber: "WBILREMOVE"}
	cacheService.Set(order.OrderUID, order)

	_, exists := cacheService.Get(order.OrderUID)
	assert.True(t, exists)

	cacheService.Remove(order.OrderUID)
	_, exists = cacheService.Get(order.OrderUID)
	assert.False(t, exists)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheService := NewCache().(*ShardedCache)

	go func() {
		for i := 0; i < 100; i++ {
			order := models.Order{
				OrderUID:    string(rune(i)),
				TrackNumber: string(rune(i)),
			}
			cacheService.Set(order.OrderUID, order)
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cacheService.Get(string(rune(i)))
		}
	}()

	time.Sleep(100 * time.Millisecond)

	assert.True(t, len(cacheService.GetAll()) > 0)
}
