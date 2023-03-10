package gee

import (
	"log"
	"net/http"
	"strings"
)

//HandlerFunc defines the request handler user gee 		定义HandlerFunc函数  用于处理请求
type HandlerFunc func(*Context)

//RouterGroup  定义RouterGroup结构体  用于定义路由组
type RouterGroup struct {
	prefix     string        //前缀
	middleware []HandlerFunc //中间件
	parent     *RouterGroup  //父类路由组   支持嵌套
	engine     *Engine       //所有的路由组都共享一个engine实例
}

//Engine implement the interface of servehttp 		定义engine结构体  实现了servehttp接口
type Engine struct {
	*RouterGroup                //继承RouterGroup
	router       *router        //定义一个router实例
	groups       []*RouterGroup // store all groups 存储所有的路由组
}

//New is the Constructor of gee.engine 		定义New函数  用于创建一个engine实例
func New() *Engine {
	engine := &Engine{router: newRouter()}             //创建一个engine实例
	engine.RouterGroup = &RouterGroup{engine: engine}  //初始化一个RouterGroup
	engine.groups = []*RouterGroup{engine.RouterGroup} //初始化groups
	return engine
}

// Group 组的定义是为了创建一个新的RouterGroup
// 记住所有组都共享同一个engine实例
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// addRoute is a private method to add route to the router 定义addRoute函数  用于添加路由
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

//GET 定义了添加GET请求的方法
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

//POST 定义了添加POST请求的方法
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

//Run defines the method to start a http server 定义run函数  用于启动http服务
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// Use is used to add middleware to the group  定义use函数  用于添加中间件
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middleware = append(group.middleware, middleware...)
}

// ServeHTTP defines the interface to implement the http.Handler 定义ServeHTTP函数  实现了http.Handler接口
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	//遍历所有的路由组
	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middleware...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	engine.router.handle(c)

}
