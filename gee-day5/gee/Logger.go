package gee

import (
	"log"
	"time"
)

// Logger 定义Logger函数  用于记录请求和响应
func Logger() HandlerFunc {
	return func(c *Context) {
		// Start timer 开始计时
		t := time.Now()
		// Process request 	 处理请求 执行下一个中间件 也就是下一个路由

		/**
		i:=1
		if i==2{
			c.Next()
		}else {
			c.Fail(500,"Internal Server Error")
		}
		*/

		c.Next()
		// Calculate resolution time  计算处理时间
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
