package cache

import (
	"hash/fnv"
	"sync"

	"l0/internal/models"
)

type Cache interface {
	Set(uid string, order models.Order)
	Get(uid string) (models.Order, bool)
	GetAll() map[string]models.Order
	Remove(uid string)
}

type shard struct {
	mu     sync.RWMutex
	orders map[string]models.Order
}

type ShardedCache struct {
	shards []shard
}

const numShards = 32

func NewCache() Cache {
	shards := make([]shard, numShards)
	for i := range shards {
		shards[i].orders = make(map[string]models.Order)
	}
	return &ShardedCache{shards: shards}
}

func (c *ShardedCache) getShard(uid string) *shard {
	h := fnv.New32a()
	h.Write([]byte(uid))
	return &c.shards[h.Sum32()%uint32(numShards)]
}

func (c *ShardedCache) Set(uid string, order models.Order) {
	s := c.getShard(uid)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[uid] = order
}

func (c *ShardedCache) Get(uid string) (models.Order, bool) {
	s := c.getShard(uid)
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[uid]
	return order, ok
}

func (c *ShardedCache) Remove(uid string) {
	s := c.getShard(uid)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.orders, uid)
}

func (c *ShardedCache) GetAll() map[string]models.Order {
	result := make(map[string]models.Order)

	for i := range c.shards {
		s := &c.shards[i]
		s.mu.RLock()
		for k, v := range s.orders {
			result[k] = v
		}
		s.mu.RUnlock()
	}

	return result
}
