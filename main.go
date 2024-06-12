package main

import (
	"context"
	"fmt"
	"github.com/advanced-go/resource-host/register"
	"github.com/advanced-go/stdlib/core"
	"github.com/advanced-go/stdlib/host"
	"github.com/advanced-go/stdlib/httpx"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

const (
	portKey                 = "PORT"
	addr                    = "0.0.0.0:8081"
	writeTimeout            = time.Second * 300
	readTimeout             = time.Second * 15
	idleTimeout             = time.Second * 60
	healthLivelinessPattern = "/health/liveness"
	healthReadinessPattern  = "/health/readiness"
)

func main() {
	//os.Setenv(portKey, "0.0.0.0:8082")
	port := os.Getenv(portKey)
	if port == "" {
		port = addr
	}
	start := time.Now()
	displayRuntime(port)
	handler, ok := startup(http.NewServeMux())
	if !ok {
		os.Exit(1)
	}
	fmt.Println(fmt.Sprintf("started : %v", time.Since(start)))
	srv := http.Server{
		Addr: port,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      handler,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		} else {
			log.Printf("HTTP server Shutdown")
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}

func displayRuntime(port string) {
	fmt.Printf("addr    : %v\n", port)
	fmt.Printf("vers    : %v\n", runtime.Version())
	fmt.Printf("os      : %v\n", runtime.GOOS)
	fmt.Printf("arch    : %v\n", runtime.GOARCH)
	fmt.Printf("cpu     : %v\n", runtime.NumCPU())
	fmt.Printf("env     : %v\n", core.EnvStr())
}

func startup(r *http.ServeMux) (http.Handler, bool) {
	// Initialize logging
	register.Logging()

	// Initialize host configuration
	if !register.Configuration() {
		return r, false
	}

	// Initialize host Exchange
	host.SetHostTimeout(time.Second * 3)
	host.SetAuthExchange(AuthHandler, nil)
	err := register.IngressExchange()
	if err != nil {
		log.Printf(err.Error())
		return r, false
	}

	// Initialize HTTP controllers
	err = register.EgressController()
	if err != nil {
		log.Printf(err.Error())
		return r, false
	}

	// Initialize health handlers
	r.Handle(healthLivelinessPattern, http.HandlerFunc(healthLivelinessHandler))
	r.Handle(healthReadinessPattern, http.HandlerFunc(healthReadinessHandler))

	// Route all other requests to host proxy
	r.Handle("/", http.HandlerFunc(host.HttpHandler))
	return r, true
}

func healthLivelinessHandler(w http.ResponseWriter, r *http.Request) {
	var status = core.StatusOK()
	if status.OK() {
		httpx.WriteResponse[core.Log](w, nil, status.HttpCode(), []byte("up"), nil)
	} else {
		httpx.WriteResponse[core.Log](w, nil, status.HttpCode(), nil, nil)
	}
}

func healthReadinessHandler(w http.ResponseWriter, r *http.Request) {
	var status = core.StatusOK()
	if status.OK() {
		httpx.WriteResponse[core.Log](w, nil, status.HttpCode(), []byte("up"), nil)
	} else {
		httpx.WriteResponse[core.Log](w, nil, status.HttpCode(), nil, nil)
	}
}

func AuthHandler(r *http.Request) (*http.Response, *core.Status) {
	/*
		if r != nil {
			tokenString := r.Header.Get(host.Authorization)
			if tokenString == "" {
				status := core.NewStatus(http.StatusUnauthorized)
				return &http.Response{StatusCode: status.HttpCode()}, status
				//w.WriteHeader(http.StatusUnauthorized)
				//fmt.Fprint(w, "Missing authorization header")
			}
		}
	*/
	return &http.Response{StatusCode: http.StatusOK}, core.StatusOK()

}
