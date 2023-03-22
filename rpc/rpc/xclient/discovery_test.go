package xclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	oldServers = []string{"tcp@[::]:46059", "tcp@[::]:46060"}
	servers    = []string{"tcp@[::]:46061", "tcp@[::]:46062", "tcp@[::]:46063"}
)

func TestRefresh(t *testing.T) {
	// TODO
}

func TestUpdate(t *testing.T) {
	d := NewMultiServerDiscovery(oldServers)
	d.Update(servers)
	require.EqualValues(t, servers, d.servers)
}

func TestMultiServerDiscoveryGet(t *testing.T) {
	d := NewMultiServerDiscovery(servers)
	server, err := d.Get(RandomSelect)
	require.NoError(t, err)
	require.Contains(t, servers, server)
	d.index = 0
	server, err = d.Get(RoundRobinSelect)
	require.NoError(t, err)
	require.Equal(t, servers[0], server)
	server, err = d.Get(RoundRobinSelect)
	require.NoError(t, err)
	require.Equal(t, servers[1], server)
	server, err = d.Get(RoundRobinSelect)
	require.NoError(t, err)
	require.Equal(t, servers[2], server)
	server, err = d.Get(RoundRobinSelect)
	require.NoError(t, err)
	require.Equal(t, servers[0], server)
}

func TestMultiServerDiscoveryGetAll(t *testing.T) {
	d := NewMultiServerDiscovery(servers)
	all, err := d.GetAll()
	require.NoError(t, err)
	require.EqualValues(t, servers, all)
}
