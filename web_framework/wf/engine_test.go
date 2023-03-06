package wf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestEngine() *Engine {
	r := New()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	require.Equal(t, parsePath("/"), []string{})
	require.Equal(t, parsePath("/p/:name"), []string{"p", ":name"})
	require.Equal(t, parsePath("/:name"), []string{":name"})
	require.Equal(t, parsePath("/p/*"), []string{"p", "*"})
	require.Equal(t, parsePath("/*"), []string{"*"})
	require.Equal(t, parsePath("/p/*name/*"), []string{"p", "*name"})
}

func TestGetRoute(t *testing.T) {
	r := newTestEngine()
	n, params := r.getRoute("GET", "/hello/myname")
	require.NotNil(t, n)
	require.Equal(t, "/hello/:name", n.path)
	require.Equal(t, "myname", params["name"])

	n, params = r.getRoute("GET", "/assets/images/logo.png")
	require.NotNil(t, n)
	require.Equal(t, "/assets/*filepath", n.path)
	require.Equal(t, "images/logo.png", params["filepath"])
}

func TestMixPath(t *testing.T) {
	r := New()

	r.addRoute("GET", "/:name", nil)
	r.addRoute("GET", "/a", nil)
	require.Panics(t, func() { r.addRoute("GET", "/a", nil) }, "Duplicate routing")
	require.Panics(t, func() { r.addRoute("GET", "/:fix", nil) }, "Duplicate routing")
	r.addRoute("GET", "/*path", nil)

	n, p := r.getRoute("GET", "/a")
	t.Log(n.path)
	require.Empty(t, p)

	n, _ = r.getRoute("GET", "/b")
	t.Log(n.path)

	r.addRoute("GET", "/:last/ok/*path", nil)

	n, p = r.getRoute("GET", "/c/ok/cute.jpg")
	t.Log(n.path)
	t.Log(p["last"])
	t.Log(p["path"])

	n, p = r.getRoute("GET", "/d/image.png")
	t.Log(n.path)
	t.Log(p["path"])

	// require.Nil(t, n)
}
