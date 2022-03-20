package rpc

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
)

type cache struct {
	dauPool *lru.Cache
	txPool  *lru.Cache
	mutex   sync.Mutex
}

func NewCache() *cache {
	dauPool, _ := lru.New(1024 * 1024)
	txPool, _ := lru.New(1024 * 1024)
	c := &cache{
		dauPool: dauPool,
		txPool:  txPool,
	}
	return c
}

func (c *cache) getDau(key string) (count uint64, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if dau, ok := c.dauPool.Get(key); ok {
		return dau.(uint64), nil
	}
	return 0, ErrNoCahceItem
}

func (c *cache) updateDau(key string, count uint64) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dauPool.Add(key, count)
	return
}

func (c *cache) getTxCount(key string) (count uint64, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if txCount, ok := c.txPool.Get(key); ok {
		return txCount.(uint64), nil
	}
	return 0, ErrNoCahceItem
}

func (c *cache) updateTxCount(key string, count uint64) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.txPool.Add(key, count)
	return
}
