package go_nordigen

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

type reqAuditor struct{}

func (a *reqAuditor) ID() string {
	return time.Now().Format("2006-01-02T15:04:05")
}

func (a *reqAuditor) Request(id string, req *http.Request) {
	r, _ := httputil.DumpRequestOut(req, true)

	fmt.Println(id, "request", string(r))
}
func (a *reqAuditor) Response(id string, res *http.Response) {
	r, _ := httputil.DumpResponse(res, true)

	fmt.Println(id, "response", string(r))
}
