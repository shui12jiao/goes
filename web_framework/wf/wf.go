package wf

import (
	"net/http"
	"strings"
	"text/template"
)

type HandlerFunc func(*Context)

type HandlersChain []HandlerFunc

type Engine struct {
	RouterGroup
	roots         map[string]*node //method to root
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

func New() *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			prefix:   "/",
			handlers: nil,
		},
		roots: make(map[string]*node),
	}
	engine.RouterGroup.engine = engine
	return engine
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r, engine)
	engine.handle(c)
}

func (engine *Engine) Run(address string) error {
	return http.ListenAndServe(address, engine)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) Use(middlewares ...HandlerFunc) *Engine {
	engine.RouterGroup.Use(middlewares...)
	return engine
}

func (engine *Engine) addRoute(method string, path string, handlers HandlersChain) {
	parts := parsePath(path)
	root, ok := engine.roots[method]
	if !ok {
		engine.roots[method] = &node{}
		root = engine.roots[method]
	}
	n := root.search(parts, 0)
	if n != nil {
		//匹配 判断优先级是否重复
		// eg:?
		// /h1/:name -> /h1/name yes
		// /h1/:name -> /h1/*path yes
		// /h1/:name -> /h1/:id no
		// /h1/:name1/name2 -> /h1/name2/:name2 no? TODO
		parts := parsePath(path)       //插入
		pathParts := parsePath(n.path) //已有
		if len(pathParts) != len(parts) {
			goto insert
		}
		//长度相同
		for index, part := range parts {
			comPart := pathParts[index]
			if part == comPart {
				continue
			} else if comPart[0] != part[0] {
				goto insert
			} else {
				panic("Duplicate routing")
			}
		}
		panic("Duplicate routing")
	}
insert:
	engine.roots[method].insert(path, parts, handlers, 0)
}

func (engine *Engine) getRoute(method string, path string) (n *node, params map[string]string) {
	root, ok := engine.roots[method]
	if !ok {
		return nil, nil
	}

	searchParts := parsePath(path)
	n = root.search(searchParts, 0)
	if n == nil {
		return nil, nil
	}

	params = make(map[string]string)
	parts := parsePath(n.path)
	for index, part := range parts {
		if part[0] == ':' {
			params[part[1:]] = searchParts[index]
		} else if part[0] == '*' {
			params[part[1:]] = strings.Join(searchParts[index:], "/")
			break
		}
	}
	return n, params
}

func (engine *Engine) handle(c *Context) {
	n, params := engine.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		c.handlers = n.handlers
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
	c.Next()
}

func parsePath(path string) []string {
	vs := strings.Split(path, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}
