package http

import (
	"net/http"
	"strconv"
	"time"
)

const (
	_httpHeaderTimeout = "X-Request-Timeout"
	_httpHeaderTraceId = "X-Trace-Id"
)

// get timeout from request header
// similar as grpc
func getTimeout(req *http.Request) time.Duration {
	to := req.Header.Get(_httpHeaderTimeout)
	timeout, err := strconv.ParseInt(to, 10, 64)
	if err == nil && timeout > 20 {
		timeout -= 20 // reduce 20ms every time.
	}
	return time.Duration(timeout) * time.Millisecond
}

// setTimeout set timeout into http request.
func setTimeout(req *http.Request, timeout time.Duration) {
	td := int64(timeout / time.Millisecond)
	req.Header.Set(_httpHeaderTimeout, strconv.FormatInt(td, 10))
}
