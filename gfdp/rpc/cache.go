package rpc

import "sync"

type cache struct {
	dauPool map[string]uint64
	txPool  map[string]uint64
	mutex   sync.Mutex
}

func NewCache() *cache {
	c := &cache{
		dauPool: make(map[string]uint64),
		txPool:  make(map[string]uint64),
	}
	return c
}

func (c *cache) getDau(key string) (count uint64, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if dau, ok := c.dauPool[key]; ok {
		return dau, nil
	}
	return 0, ErrNoCahceItem
}

func (c *cache) updateDau(key string, count uint64) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dauPool[key] = count
	return
}

func (c *cache) getTxCount(key string) (count uint64, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if txCount, ok := c.txPool[key]; ok {
		return txCount, nil
	}
	return 0, ErrNoCahceItem
}

func (c *cache) updateTxCount(key string, count uint64) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.txPool[key] = count
	return
}
