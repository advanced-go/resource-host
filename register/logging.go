package register

import (
	"fmt"
	"github.com/advanced-go/stdlib/access"
	"github.com/advanced-go/stdlib/core"
	fmt2 "github.com/advanced-go/stdlib/fmt"
	"github.com/advanced-go/stdlib/httpx"
	"github.com/advanced-go/stdlib/uri"
	"time"
)

func Logging() {
	// Override access logger
	access.SetLogFn(logger)
}

func logger(o core.Origin, traffic string, start time.Time, duration time.Duration, req any, resp any, from, routeName, routeTo string, timeout time.Duration, rateLimit float64, rateBurst int, reasonCode string) {
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
		"\"rate-limit\":%v, "+
		"\"rate-burst\":%v, "+
		"\"rc\":%v }",
		//fmt2.JsonString(o.Region),
		//fmt2.JsonString(o.Zone),
		//fmt2.JsonString(o.SubZone),
		//fmt2.JsonString(o.App),
		//fmt2.JsonString(o.InstanceId),

		traffic,
		fmt2.FmtRFC3339Millis(start),
		access.Milliseconds(duration),

		fmt2.JsonString(newReq.Header.Get(httpx.XRequestId)),
		//fmt2.JsonString(req.Header.Get(httpx.XRelatesTo)),
		//fmt2.JsonString(req.Proto),
		fmt2.JsonString(newReq.Method),
		fmt2.JsonString(o.Host),
		fmt2.JsonString(from),
		fmt2.JsonString(access.CreateTo(newReq)),
		fmt2.JsonString(url),
		fmt2.JsonString(parsed.Query),

		//fmt2.JsonString(path),

		newResp.StatusCode,
		fmt.Sprintf("%v", newResp.ContentLength),
		fmt2.JsonString(access.Encoding(newResp)),

		fmt2.JsonString(routeName),
		//fmt2.JsonString(routeTo),
		access.Milliseconds(timeout),
		fmt.Sprintf("%v", rateLimit),
		rateBurst,
		fmt2.JsonString(reasonCode),
	)
	fmt.Printf("%v\n", s)
	//return s
}
