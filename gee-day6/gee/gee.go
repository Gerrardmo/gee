package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
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
	*RouterGroup                     //继承RouterGroup
	router        *router            //定义一个router实例
	groups        []*RouterGroup     // store all groups 存储所有的路由组
	htmlTemplates *template.Template // for html render 用于html渲染 模板 有点类似于jsp
	funcMap       template.FuncMap   // for html render 用于html渲染 函数映射 自定义函数 例如：{{now}}
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
	c.engine = engine
	engine.router.handle(c)

}

//createStaticHandler  定义静态文件处理函数
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.prefix + relativePath
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

//Static serve static files 定义静态文件处理函数
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// register GET handlers
	group.GET(urlPattern, handler)
}

//SetFuncMap 用于设置模板 例如：engine.SetHTMLTemplate(template)
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	//这里是设置模板 例如：engine.SetHTMLTemplate(template)
	engine.funcMap = funcMap
}

//LoadHTMLGlob 用于加载模板 例如：engine.LoadHTMLGlob("templates/*")
func (engine *Engine) LoadHTMLGlob(pattern string) {
	//这里是加载模板 例如：engine.LoadHTMLGlob("templates/*") 会加载templates目录下的所有模板 例如：templates/index.html
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
