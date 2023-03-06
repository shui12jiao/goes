package dc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var db = map[string]string{
	"Alice":   "123",
	"Bob":     "456",
	"Charlie": "789",
}

func TestGroupGet(t *testing.T) {
	dbHits := make(map[string]int, len(db))
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			v, ok := db[key]
			if ok {
				dbHits[key]++
				return []byte(v), nil
			}
			return nil, errors.New("no such key")
		},
	))

	g := GetGroup("scores")
	val, err := g.Get("")
	require.Empty(t, val)
	require.EqualError(t, err, "key is required")

	val, err = g.Get("Alice")
	require.Equal(t, db["Alice"], val.String())
	require.NoError(t, err)
	require.Equal(t, 1, dbHits["Alice"])

	val, err = g.Get("Bob")
	require.Equal(t, db["Bob"], val.String())
	require.NoError(t, err)
	require.Equal(t, 1, dbHits["Bob"])

	val, err = g.Get("Charlie")
	require.Equal(t, db["Charlie"], val.String())
	require.NoError(t, err)
	require.Equal(t, 1, dbHits["Charlie"])

	val, err = g.Get("David")
	require.Empty(t, val)
	require.EqualError(t, err, "no such key")

	val, err = g.Get("Alice")
	require.Equal(t, db["Alice"], val.String())
	require.NoError(t, err)
	require.Equal(t, 1, dbHits["Alice"])
}
