// Package strcache implements a simple LFU cache.
package strcache

import (
	"errors"
	"sync"

	"github.com/ZYunH/strcache/internal/skiplist"
)

// ByteSize represents the size of cache. As its name indicates, the unit is Byte.
type ByteSize float64

// Shortcut for cache size.
const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// Exported errors.
var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrNoEnoughSpace = errors.New("no enough space")
)

// Cache represents a LFU cache with max size limit.
type Cache struct {
	m       map[string]*skiplist.Node
	nowsize ByteSize
	maxsize ByteSize
	skl     *skiplist.SkipList

	mu sync.Mutex
}

// New returns a new LFU cache with max size.
func New(maxsize ByteSize) *Cache {
	return &Cache{
		m:       make(map[string]*skiplist.Node, 10),
		maxsize: maxsize,
		skl:     skiplist.NewDefault(),
	}
}

// Set set a key associated with the val. The error occurred only if the size of
// the input value greater than the max size.
func (c *Cache) Set(key string, val string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if node, ok := c.m[key]; ok {
		if node.Val == val {
			return nil
		}
		// Delete the key.
		c.skl.Delete(node.Score, node.Key, node.LastAccess)
		c.nowsize -= ByteSize(len(node.Val))
		delete(c.m, key)
	}
	// Insert a new node.
	size := ByteSize(len(val))
	if size > c.maxsize {
		return ErrNoEnoughSpace
	} else if size > (c.maxsize - c.nowsize) {
		// Delete some nodes for more space.
		morespace := size - (c.maxsize - c.nowsize)
		var removeNodes []*skiplist.Node
		x := c.skl.Head()
		for x != nil && morespace > 0 {
			morespace -= ByteSize(len(x.Val))
			removeNodes = append(removeNodes, x)
			x = x.Next()
		}
		for _, dn := range removeNodes {
			delete(c.m, dn.Key)
			c.skl.Delete(dn.Score, dn.Key, dn.LastAccess)
			c.nowsize -= ByteSize(len(dn.Val))
		}
	}
	// We have enough space to store this key.
	newNode := c.skl.Insert(1, key, val)
	c.m[key] = newNode
	c.nowsize += ByteSize(len(val))
	return nil
}

// Get get a value associated with the input key.
// If the key is not in the cache, ErrKeyNotFound will returned.
func (c *Cache) Get(key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n, ok := c.m[key]; ok {
		c.skl.Update(n.Score, n.Key, n.LastAccess, n.Score+1)
		return n.Val, nil
	}
	return "", ErrKeyNotFound
}

// Del deletes a key in the cache.
// If the key is not in the cache, ErrKeyNotFound will returned.
func (c *Cache) Del(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if node, ok := c.m[key]; ok {
		c.skl.Delete(node.Score, node.Key, node.LastAccess)
		c.nowsize -= ByteSize(len(node.Val))
		delete(c.m, key)
		return nil
	}
	return ErrKeyNotFound
}

// Len returns the length of cache.
func (c *Cache) Len() int {
	c.mu.Lock()
	x := len(c.m)
	c.mu.Unlock()
	return x
}
