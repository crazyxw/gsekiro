package main

import (
	"fmt"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func VerifySignMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// 签名校验
		reqParams := request.URL.Query()
		fmt.Println("收到请求", request.URL.String())
		var vkey = reqParams.Get("vkey")
		if vKey != "" && vkey != vKey {
			writer.Write([]byte("签名有误"))
			return
		}
		handler.ServeHTTP(writer, request)
	})
}

// 接口测时
func MetricMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			fmt.Printf("path:%s elapsed:%fs\n", r.URL.Path, time.Since(start).Seconds())
		}()
		handler.ServeHTTP(w, r)
	})
}

func applyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
