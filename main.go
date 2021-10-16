package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

/*
	1. 接收客户端request，并将request中带的header写入response header
	2. 读取当前系统的环境变量中VERSION配置，并写入response header
	3. Server端记录访问日志包括客户端IP，HTTP返回码，输出到Server端的标准输出
	4. 当访问localhost/healthz时，应返回200
*/

type newResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterFunc(w http.ResponseWriter) *newResponseWriter {
	return &newResponseWriter{w, http.StatusOK}
}

func (nrw *newResponseWriter) WriteHeader(code int) {
	nrw.statusCode = code
	nrw.ResponseWriter.WriteHeader(code)
}

func newHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 接收客户端request，并将request中带的header写入response header
		headers := r.Header
		if len(headers) > 0 {
			for k, v := range headers {
				w.Header().Set(k, v[0])
			}
		}

		// 读取当前系统的环境变量中VERSION配置，并写入response header
		envValue := os.Getenv("VERSION")
		if envValue == "" {
			envValue = "unknown"
		}
		w.Header().Set("version", envValue)

		// Server端记录访问日志包括客户端IP，HTTP返回码，输出到Server端的标准输出
		nrw := newResponseWriterFunc(w)
		handler.ServeHTTP(nrw, r)
		statusCode := nrw.statusCode
		remoteAddr := r.RemoteAddr
		lastIndex := strings.LastIndex(remoteAddr, ":")
		clientIP := remoteAddr[:lastIndex]
		log.Printf("%s %s %s %d %s", clientIP, r.Method, r.URL.Path, statusCode, http.StatusText(statusCode))
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		io.WriteString(w, "<h1>Home page</h1>")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed.")
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// 当GET访问localhost/healthz时，返回200
	if r.Method == http.MethodGet {
		io.WriteString(w, "200")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed.")
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, `<h1 style="color: red">Page not found.</h1>`)
}

func main() {
	http.Handle("/", newHandler(http.HandlerFunc(notFoundHandler)))
	http.Handle("/index", newHandler(http.HandlerFunc(indexHandler)))
	http.Handle("/healthz", newHandler(http.HandlerFunc(healthzHandler)))
	err := http.ListenAndServe("localhost:80", nil)
	if err != nil {
		log.Fatal(err)
	}
}