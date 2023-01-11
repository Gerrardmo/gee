package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
		roots:    make(map[string]*node),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

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

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

//getRoute 用于查找路由
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]
	//如果没有找到对应的method，直接返回
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	//如果没有找到对应的路由，直接返回
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			//如果是动态路由
			if part[0] == ':' {
				//将参数名作为key，参数值作为value
				params[part[1:]] = searchParts[index]
			}
			//如果是通配符路由
			if part[0] == '*' && len(part) > 1 {
				//将通配符后面的参数保存到params中
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	//从context中获取请求方法和请求路径
	n, params := r.getRoute(c.Method, c.Path)
	//如果没有找到对应的路由，直接返回
	if n != nil {
		key := c.Method + "-" + n.pattern
		c.Params = params
		//执行对应的handler
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		//如果没有找到对应的路由，直接返回404
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s", c.Path)
		})
	}
	c.Next()
}
