package dc

import (
	"dc/lru"
	"dc/singleflight"
	"errors"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	maxBytes  int64
	picker    PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mu.Lock()
	defer mu.Unlock()

	_, exits := groups[name]
	if exits {
		panic("group already exits")
	}

	g := &Group{
		name:     name,
		getter:   getter,
		maxBytes: maxBytes,
		loader:   &singleflight.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) RegisterPeerPicker(picker PeerPicker) {
	if g.picker != nil {
		panic("RegisterPeerPicker is called repeatedly")
	}

	g.picker = picker
}

func (g *Group) Get(key string) (ByteView, error) {
	if len(key) == 0 {
		return ByteView{}, errors.New("key is required")
	}

	val, ok := g.mainCache.get(key)
	if ok {
		return val, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (val ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.picker != nil {
			peer, ok := g.picker.PickPeer(key)
			if ok {
				val, err = g.getFromPeer(peer, key)
				if err == nil {
					return val, nil
				}
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		val = viewi.(ByteView)
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	val := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, val)
	return val, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
// 	req := &pb.Request{
// 		Group: g.name,
// 		Key:   key,
// 	}
// 	res := &pb.Response{}
// 	err := peer.Get(req, res)
// 	if err != nil {
// 		return ByteView{}, err
// 	}
// 	return ByteView{b: res.Value}, nil
// }

func (g *Group) populateCache(key string, val ByteView) {
	g.mainCache.add(key, val)
}

// Cache
type cache struct {
	nBytes uint64
	mu     sync.Mutex
	lru    *lru.Cache
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(0)
	}
	c.lru.Add(key, value)
	c.nBytes += uint64(value.Len()) + uint64(len(key))
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return ByteView{}, false
	}
	value, ok := c.lru.Get(key)
	if !ok {
		return ByteView{}, false
	}
	return value.(ByteView), true
}
