package main

import (
	"cloud-native-study/httpserver/metrics"
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
		glog.Infof("%s %s %s %d %s", clientIP, r.Method, r.URL.Path, statusCode, http.StatusText(statusCode))
	})
}

func randInt(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(4).Info("Entering index handler")
	timer := metrics.NewTimer()
	defer timer.ObserveTotal()
	delay := randInt(1000, 2000)
	fmt.Println(delay)
	time.Sleep(time.Millisecond * time.Duration(delay))
	if r.Method == http.MethodGet {
		io.WriteString(w, "<h1>Home page</h1>")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed.")
	}
	glog.V(4).Infof("Respond in %d ms", delay)
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
	flag.Set("v", "4")
	glog.V(2).Info("Starting http server...")
	metrics.Register()
	addr := flag.String("server addr", "0.0.0.0:80", "server address")
	mux := http.NewServeMux()
	mux.Handle("/", newHandler(http.HandlerFunc(notFoundHandler)))
	mux.Handle("/index", newHandler(http.HandlerFunc(indexHandler)))
	mux.Handle("/healthz", newHandler(http.HandlerFunc(healthzHandler)))
	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    *addr,
		Handler: mux,
	}
	// make sure idle connections returned
	processed := make(chan struct{})
	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); nil != err {
			glog.Fatalf("server shutdown failed, err: %v", err)
		}
		glog.Info("server gracefully shutdown.")

		close(processed)
	}()

	// server
	glog.Infof("httpserver started on %v.", *addr)
	err := server.ListenAndServe()
	if http.ErrServerClosed != err {
		glog.Fatalf("server not gracefully shutdown, err: %v", err)
	}

	// waiting for goroutine above processed
	<-processed
}
