package lru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	cache := New(3)
	cache.OnEvicted = func(key Key, value Value) {
		t.Logf("key %s evicted\n", key)
	}
	cache.Add("a", 1)
	cache.Add("b", 2)
	cache.Add("c", 3)
	cache.Add("d", 4)

	e, ok := cache.Get("a")
	require.False(t, ok)
	require.Nil(t, e)
	e, ok = cache.Get("b")
	require.True(t, ok)
	require.Equal(t, 2, e)
	e, ok = cache.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, e)
	e, ok = cache.Get("d")
	require.True(t, ok)
	require.Equal(t, 4, e)

	cache.RemoveOldest()
	e, ok = cache.Get("b")
	require.False(t, ok)
	require.Nil(t, e)

	cache.Remove("d")
	e, ok = cache.Get("d")
	require.False(t, ok)
	require.Nil(t, e)

	len := cache.Len()
	require.Equal(t, 1, len)

	cache.Clear()
	len = cache.Len()
	require.Equal(t, 0, len)
	require.Nil(t, cache.cache)
	require.Nil(t, cache.list)
}
