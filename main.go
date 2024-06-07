package main

import (
	"context"
	"fmt"
	guidehttp "github.com/advanced-go/guidance/http"
	guidemod "github.com/advanced-go/guidance/module"
	observhttp "github.com/advanced-go/observation/http"
	observmod "github.com/advanced-go/observation/module"
	searchhttp "github.com/advanced-go/search/http"
	searchmod "github.com/advanced-go/search/module"
	"github.com/advanced-go/stdlib/access"
	"github.com/advanced-go/stdlib/core"
	fmt2 "github.com/advanced-go/stdlib/fmt"
	"github.com/advanced-go/stdlib/host"
	"github.com/advanced-go/stdlib/httpx"
	"github.com/advanced-go/stdlib/uri"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
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
	// Override access logger
	access.SetLogFn(logger)

	// Run host startup where all registered resources/packages will be sent a startup configuration message
	m := createPackageConfiguration()
	if !host.Startup(time.Second*4, m) {
		return r, false
	}

	// Initialize host Exchange
	host.SetHostTimeout(time.Second * 3)
	host.SetAuthExchange(AuthHandler, nil)
	err := registerExchanges()
	if err != nil {
		log.Printf(err.Error())
		return r, false
	}

	// Initialize HTTP controllers
	err = registerControllers()
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

// TO DO : create package configuration information for startup
func createPackageConfiguration() host.ContentMap {
	return make(host.ContentMap)
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

func logger(o core.Origin, traffic string, start time.Time, duration time.Duration, req any, resp any, routeName, routeTo string, timeout time.Duration, rateLimit float64, rateBurst int, reasonCode string) {
	newReq := access.BuildRequest(req)
	newResp := access.BuildResponse(resp)
	url, parsed := uri.ParseURL(newReq.Host, newReq.URL)
	o.Host = access.Conditional(o.Host, parsed.Host)

	s := fmt.Sprintf("{"+
		//"\"region\":%v, "+
		//"\"zone\":%v, "+
		//"\"sub-zone\":%v, "+
		//"\"instance-id\":%v, "+
		"\"traffic\":\"%v\", "+
		"\"start\":%v, "+
		"\"duration\":%v, "+
		"\"request-id\":%v, "+
		//"\"relates-to\":%v, "+
		//"\"proto\":%v, "+
		"\"method\":%v, "+
		"\"host\":%v, "+
		"\"from\":%v, "+
		"\"to\":%v, "+
		"\"uri\":%v, "+
		"\"query\":%v, "+
		//"\"path\":%v, "+
		"\"status-code\":%v, "+
		"\"bytes\":%v, "+
		"\"encoding\":%v, "+
		"\"route\":%v, "+
		//"\"route-to\":%v, "+
		"\"timeout\":%v, "+
		//"\"timeout\":%v, "+
		//"\"timeout\":%v, "+
		"\"rc\":%v }",
		//fmt2.JsonString(o.Region),
		//fmt2.JsonString(o.Zone),
		//fmt2.JsonString(o.SubZone),
		//fmt2.JsonString(o.App),
		//fmt2.JsonString(o.InstanceId),

		traffic,
		fmt2.FmtRFC3339Millis(start),
		strconv.Itoa(int(duration/time.Duration(1e6))),

		fmt2.JsonString(newReq.Header.Get(httpx.XRequestId)),
		//fmt2.JsonString(req.Header.Get(httpx.XRelatesTo)),
		//fmt2.JsonString(req.Proto),
		fmt2.JsonString(newReq.Method),
		fmt2.JsonString(o.Host),
		fmt2.JsonString(newReq.Header.Get(core.XAuthority)),
		fmt2.JsonString(uri.UprootAuthority(newReq.URL)),
		fmt2.JsonString(url),
		fmt2.JsonString(parsed.Query),

		//fmt2.JsonString(path),

		newResp.StatusCode,
		fmt.Sprintf("%v", newResp.ContentLength),
		fmt2.JsonString(access.Encoding(newResp)),

		fmt2.JsonString(routeName),
		//fmt2.JsonString(routeTo),
		int(timeout/time.Duration(1e6)),
		fmt2.JsonString(reasonCode),
	)
	fmt.Printf("%v\n", s)
	//return s
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

func registerExchanges() error {
	err := host.RegisterExchange(searchmod.Authority, host.NewAccessLogIntermediary(searchmod.RouteName, searchhttp.Exchange))
	if err != nil {
		return err
	}
	err = host.RegisterExchange(guidemod.Authority, host.NewAccessLogIntermediary(guidemod.RouteName, guidehttp.Exchange))
	if err != nil {
		return err
	}
	err = host.RegisterExchange(observmod.Authority, host.NewAccessLogIntermediary(observmod.RouteName, observhttp.Exchange))
	if err != nil {
		return err
	}
	return nil
}

func registerControllers() error {
	//ctrl := searchhttp.
	//for _, ctrl := range searchhttp.Controllers() {
	//	err := controller.RegisterController(ctrl)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}
