package lru

import (
	"container/list"
)

type Cache struct {
	MaxEntries int

	OnEvicted func(key Key, value Value)

	list  *list.List
	cache map[Key]*list.Element
}

type entry struct {
	key   Key
	value Value
}

type Key interface{}
type Value interface {
	// Len() int
}

func New(m int) *Cache {
	return &Cache{
		MaxEntries: m, // 0表示不限制元素个数
		OnEvicted:  nil,
		list:       list.New(), // 双向链表
		cache:      make(map[Key]*list.Element),
	}
}

func (c *Cache) Add(k Key, v Value) {
	if c.cache == nil || c.list == nil {
		c.cache = make(map[Key]*list.Element)
		c.list = list.New()
	}

	existE, ok := c.cache[k]
	if ok {
		c.list.MoveToFront(existE)
		existE.Value.(*entry).value = v
		return
	}

	if c.MaxEntries > 0 && c.list.Len() >= c.MaxEntries {
		c.RemoveOldest()
	}

	e := c.list.PushFront(&entry{key: k, value: v})
	c.cache[k] = e
}

func (c *Cache) Get(k Key) (Value, bool) {
	if c.cache == nil {
		return nil, false
	}

	e, ok := c.cache[k]
	if !ok {
		return nil, false
	}

	c.list.MoveToFront(e)
	return e.Value.(*entry).value, true
}

func (c *Cache) Remove(k Key) {
	if c.cache == nil {
		return
	}

	e, ok := c.cache[k]
	if !ok {
		return
	}

	c.removeElement(e)
}

func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}

	e := c.list.Back()
	if e != nil {
		c.removeElement(e)
	}
}

func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}

	return c.list.Len()
}

func (c *Cache) Clear() {
	if c.cache == nil {
		return
	}

	if c.OnEvicted != nil {
		for _, e := range c.cache {
			c.OnEvicted(e.Value.(*entry).key, e.Value.(*entry).value)
		}
	}

	c.list = nil
	c.cache = nil
}

func (c *Cache) SetLimit(n int) *Cache {
	c.MaxEntries = n
	return c
}

func (c *Cache) removeElement(e *list.Element) {
	c.list.Remove(e)
	ent := e.Value.(*entry)
	delete(c.cache, ent.key)
	if c.OnEvicted != nil {
		c.OnEvicted(ent.key, ent.value)
	}
}
