package customreg

import (
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// type typeId struct {
// 	dc   int64
// 	path string
// }

type cache[V any] struct {
	mu    sync.RWMutex
	num   int             // last num added ; sequence
	keys  func(e V) []any // extract UNIQUE keys for Value
	index map[any]int     // map[key](*index).num
	cache *lru.Cache[int, *index[V]]
}

type index[V any] struct {
	num   int
	keys  []any // comparable
	value V
}

func newCache[V any](indexKeys func(V) []any, size int) *cache[V] {
	c := &cache[V]{
		keys:  indexKeys,
		index: make(map[any]int),
	}
	var crit error
	c.cache, crit = lru.NewWithEvict(
		size, c.onEvictedHook,
	)
	if crit != nil {
		panic(crit)
	}
	return c
}

func (c *cache[V]) onEvictedHook(_ int, node *index[V]) {

	// c.mu.Lock()
	// defer c.mu.Unlock()

	c.del(node.num, node.keys)
}

func (c *cache[V]) del(num int, keys []any) {

	// c.mu.Lock()
	// defer c.mu.Unlock()

	var (
		ok  bool
		reg int
	)
	for _, key := range keys {
		reg, ok = c.index[key]
		if ok && reg == num {
			delete(c.index, key)
		}
	}

}

func (c *cache[V]) add(num int, keys []any) {

	// c.mu.Lock()
	// defer c.mu.Unlock()
	defer func() {
		if e := recover(); e != nil {
			err := e
			_ = err
		}
	}()

	for _, key := range keys {
		c.index[key] = num
	}

}

func (c *cache[V]) find(keys []any) (nums []int) {

	// c.mu.Lock()
	// defer c.mu.Unlock()

	var (
		ok   bool
		num  int
		e, n int
	)
	for _, key := range keys {
		if num, ok = c.index[key]; ok {
			e, n = 0, len(nums)
			for ; e < n && nums[e] != num; e++ {
				// lookup: key(s) for the same node found ?
			}
			if e < n {
				// key node found previously !
				continue
			}
			// add to the result !
			nums = append(nums, num)
		}
	}
	return // [nums] for given [keys] found !
}

func (c *cache[V]) Num() int {

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache != nil {
		return c.cache.Len()
	}

	return 0
}

func (c *cache[V]) Add(set V) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	keys := c.keys(set)  // for index fiven [set] value !
	recs := c.find(keys) // node.(index).num(s) of the [keys] found !
	// var old, new *index[V]
	// var drop []any
	var (
		node     *index[V]
		old, new []any // index keys difference !
	)

	// var num int // for this [node] !
	switch len(recs) {
	case 0:
		// Not Found (any) ! [ADD]
		// num = c.cache.Len() + 1
		// new = &index[V]{
		// 	num:   c.cache.Len() + 1,
		// 	keys:  keys,
		// 	value: set,
		// }
		c.num++ // locked(!)
		node = &index[V]{
			num:   c.num, // c.cache.Len() + 1,
			keys:  keys,
			value: set,
		}
		// old = nil
		new = keys
	case 1:
		{
			// Found (partial) ! [EDT]
			num := recs[0]
			node, _ = c.cache.Get(num) // [MUST]

			old = node.keys                     // OLD
			new = append(([]any)(nil), keys...) // copy

			node.keys = keys // NEW
			node.value = set // SET

			// [old/new] index [keys] difference
			for k, n := 0, len(old); k < n; k++ {
				for x, add := range new {
					if old[k] == add {
						new = append(new[:x], new[x+1:]...) // DO NOT [re]set !
						old = append(old[:k], old[k+1:]...) // DO NOT remove !
						n--
						k--
						break
					}
				}
			}
		}
	default:
		// some key(s) are reserved !
		return fmt.Errorf("cache: [some] key(s) reserved")
	}

	// [RE]SET ; moveToFront !
	_ = c.cache.Add(node.num, node)

	c.del(node.num, old) // OLD
	c.add(node.num, new) // NEW

	return nil
}

func (c *cache[V]) Get(key any) (reg V) {

	c.mu.RLock()
	defer c.mu.RUnlock()

	num, ok := c.index[key]
	if !ok {
		return // nil
	}

	node, _ := c.cache.Get(num)
	reg = node.value
	return // set
}

func (c *cache[V]) Del(reg V) bool {

	c.mu.Lock()
	defer c.mu.Unlock()

	keys := c.keys(reg)  // for index fiven [reg] value !
	recs := c.find(keys) // node.(index).num(s) of the [keys] found !

	switch len(recs) {
	case 0:
		// Not Found (any) !
		return false
	case 1:
		{
			// Found (partial) ! [EDT]
			num := recs[0]
			// node, _ := c.cache.Get(num) // [MUST]
			ok := c.cache.Remove(num)
			// c.onEvicted(here) ; [DEAD]LOCK !
			return ok
		}
	default:
		// Too much records found !
		// Do nothing !
		return false
	}
}

func (c *cache[V]) Range(next func(e V) bool) {
	var entries []*index[V]
	c.mu.Lock()
	if c.cache != nil {
		entries = c.cache.Values()
	}
	c.mu.Unlock()

	i, n := 0, len(entries)
	for ; i < n && next(entries[i].value); i++ {
		// iterate: next
	}
}
