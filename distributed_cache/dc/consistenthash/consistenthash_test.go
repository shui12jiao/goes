package consistenthash

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestMap(t *testing.T) {
	m := New(6, nil)
	m.Add("Alice", "Bob", "Cindy")
	require.Equal(t, 18, len(m.keys))
	require.Equal(t, "Alice", m.Get("0Alice"))
	require.Equal(t, "Alice", m.Get("1Alice"))
	require.Equal(t, "Alice", m.Get("2Alice"))
	require.Equal(t, "Bob", m.Get("0Bob"))
	require.Equal(t, "Bob", m.Get("1Bob"))
	require.Equal(t, "Bob", m.Get("2Bob"))
	require.Equal(t, "Cindy", m.Get("0Cindy"))
	require.Equal(t, "Cindy", m.Get("1Cindy"))
	require.Equal(t, "Cindy", m.Get("2Cindy"))

	hit := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := randomString(15)
		hit[m.Get(key)]++
	}
	t.Log(hit)
}

func randomString(n int) string {
	var sb strings.Builder
	k := len(characters)

	for i := 0; i < n; i++ {
		c := characters[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}
