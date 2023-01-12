package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	//origin objects
	Writer http.ResponseWriter //响应
	Req    *http.Request       //请求
	//request information
	Path   string
	Method string
	Params map[string]string //路由参数
	//response info
	StatusCode int //响应状态码
	//middleware
	handlers []HandlerFunc
	index    int
	engine   *Engine
}

//newContext 用于创建Context
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Path:   r.URL.Path,
		Method: r.Method,
		Writer: w,
		Req:    r,
		index:  -1,
	}
}

//Next 用于执行下一个中间件  执行handlefunc[]中的下一个函数
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	//如果中间件执行完毕，返回
	for ; c.index < s; c.index++ {
		//执行中间件
		c.handlers[c.index](c)
	}
}

//PostForm 用于POST 请求中获取表单参数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

//Param 用于获取路由中的参数
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

//Query 用于获取请求中的参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

//Status 设置响应状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

//SetHeader 设置响应头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

//String 用于设置响应字符串
func (c *Context) String(code int, format string, value ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, value...)))
}

//JSON 用于设置响应JSON
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	//如果编码失败，返回500错误
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

//Date 用于设置响应数据
func (c *Context) Date(code int, date []byte) {
	c.Status(code)
	c.Writer.Write(date)
}

//HTML 用于设置HTML响应
func (c *Context) HTML(code int, name string, data interface{}) {
	//设置响应头 Content-Type
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	//如果渲染失败，返回500错误,通过engine的render进行渲染
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}

//Fail 用于设置错误响应
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
