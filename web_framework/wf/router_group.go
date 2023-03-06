package wf

import (
	"net/http"
)

type RouterGroup struct {
	prefix   string
	handlers HandlersChain
	engine   *Engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	middlewares := make(HandlersChain, len(group.handlers), len(group.handlers)+1)
	copy(middlewares, group.handlers)
	return &RouterGroup{
		prefix:   joinPath(group.prefix, prefix),
		handlers: middlewares,
		engine:   group.engine,
	}
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) *RouterGroup {
	group.handlers = append(group.handlers, middlewares...)
	return group
}

func (group *RouterGroup) GET(relativePath string, handler ...HandlerFunc) {

	group.addRoute("GET", group.prefix+relativePath, handler)
}

func (group *RouterGroup) POST(relativePath string, handler ...HandlerFunc) {
	group.addRoute("POST", group.prefix+relativePath, handler)
}

func (group *RouterGroup) Static(relativePath, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	path := joinPath(relativePath, "/:filepath")
	group.GET(path, handler)
}

func (group *RouterGroup) addRoute(method string, relativePath string, handlers HandlersChain) {
	handlers = combineHandlers(group.handlers, handlers)
	group.engine.addRoute(method, relativePath, handlers)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	path := joinPath(group.prefix, relativePath)
	fileServer := http.StripPrefix(path, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func combineHandlers(former, latter HandlersChain) HandlersChain {
	handlers := make(HandlersChain, len(former)+len(latter))
	copy(handlers, former)
	copy(handlers[len(former):], latter)
	return handlers
}

func joinPath(path, newPath string) (retPath string) {
	if path[len(path)-1] == '/' {
		if newPath[0] == '/' {
			retPath = path + newPath[1:]
		} else {
			retPath = path + newPath
		}
	} else {
		if newPath[0] == '/' {
			retPath = path + newPath
		} else {
			retPath = path + "/" + newPath
		}
	}

	if retPath[len(retPath)-1] == '/' {
		retPath = retPath[:len(retPath)-1]
	}
	if retPath[0] != '/' {
		retPath = "/" + retPath
	}
	return retPath
}
