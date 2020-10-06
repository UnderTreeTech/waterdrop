package http

import (
	"net/http"
	"strconv"
	"time"
)

const (
	_contentType        = "Content-Type"
	_json               = "application/json;charset=utf-8"
	_userAgent          = "User-Agent"
	_waterdropUserAgent = "waterdrop"
	_appkey             = "Appkey"
	_appkeyLen          = 16
	_timestamp          = "Timestamp"
	_sign               = "Sign"
	_nonce              = "Nonce"
	_acceptLanguage     = "Accept-Language"
	_locale             = "zh-CN"

	_httpHeaderTimeout = "X-Request-Timeout"
	_httpHeaderTraceId = "X-Trace-Id"
)

// get timeout from request header
// similar as grpc
func getTimeout(req *http.Request) time.Duration {
	to := req.Header.Get(_httpHeaderTimeout)
	timeout, err := strconv.ParseInt(to, 10, 64)
	//reduce 10ms network transmission time for every request
	if err == nil && timeout > 5 {
		timeout -= 5
	}

	return time.Duration(timeout) * time.Millisecond
}
