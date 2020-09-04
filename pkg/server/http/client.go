package http

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/metric"

	"google.golang.org/grpc/metadata"

	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xtime"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
	tlog "github.com/opentracing/opentracing-go/log"

	"github.com/go-resty/resty/v2"
)

type ClientConfig struct {
	HostURL             string
	Timeout             time.Duration
	SlowRequestDuration time.Duration

	EnableDebug bool
	EnableSign  bool

	Key    string
	Secret string
}

type Request struct {
	URI         string
	QueryParams url.Values
	Body        interface{}
	PathParams  map[string]string
}

type Client struct {
	client *resty.Client
	config *ClientConfig
}

func NewClient(config *ClientConfig) *Client {
	cli := resty.New()

	cli.SetTimeout(config.Timeout)
	cli.SetDebug(config.EnableDebug)
	cli.SetHostURL(config.HostURL)

	return &Client{
		client: cli,
		config: config,
	}
}

func (c *Client) NewRequest(ctx context.Context, method string, req *Request, reply interface{}) (*resty.Request, error) {
	request := c.client.NewRequest()
	request.URL = req.URI
	request.Method = method
	request.SetQueryParamsFromValues(req.QueryParams)
	request.SetBody(req.Body)
	request.SetPathParams(req.PathParams)
	request.SetResult(reply)

	if c.config.EnableSign {
		ts := strconv.Itoa(int(xtime.GetCurrentUnixTime()))
		nonce := xstring.RandomString(16)
		sign, err := c.sign(ctx, method, ts, nonce, req)
		if err != nil {
			return nil, err
		}

		request.SetHeader(_appkey, c.config.Key)
		request.SetHeader(_sign, sign)
		request.SetHeader(_nonce, nonce)
		request.SetHeader(_timestamp, ts)
	}

	if method != http.MethodGet {
		request.SetHeader(_contentType, _json)
	}

	request.SetHeader(_userAgent, _waterdropUserAgent)
	request.SetHeader(_acceptLanguage, _locale)

	return request, nil
}

func (c *Client) execute(ctx context.Context, request *resty.Request) error {
	span, ctx := trace.StartSpanFromContext(ctx, request.Method+" "+request.URL)
	ext.Component.Set(span, "http")
	ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
	ext.HTTPMethod.Set(span, request.Method)
	ext.HTTPUrl.Set(span, request.URL)

	// adjust request timeout
	timeout := c.config.Timeout
	if deadline, ok := ctx.Deadline(); ok {
		derivedTimeout := time.Until(deadline)
		if timeout > derivedTimeout {
			timeout = derivedTimeout
		}
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.MD(request.Header))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer func() {
		span.Finish()
		cancel()
	}()

	request.SetHeader(_httpHeaderTimeout, strconv.Itoa(int(c.config.Timeout/1e6)))
	request.SetContext(ctx)

	trace.MetadataInjector(ctx, metadata.MD(request.Header))

	now := time.Now()
	var quota float64
	if deadline, ok := ctx.Deadline(); ok {
		quota = time.Until(deadline).Seconds()
	}

	response, err := request.Execute(request.Method, request.URL)

	estatus := status.ExtractStatus(err)
	if estatus.Code() != status.OK.Code() {
		ext.Error.Set(span, true)
		span.LogFields(tlog.String("event", "error"), tlog.Int("code", estatus.Code()), tlog.String("message", estatus.Message()))
	}

	uri := strings.TrimPrefix(request.URL, c.client.HostURL)
	metric.HTTPClientHandleCounter.Inc(uri, request.Method, c.client.HostURL, estatus.Error())
	metric.HTTPClientReqDuration.Observe(float64(time.Since(now)/time.Millisecond), uri, request.Method, c.client.HostURL)

	duration := time.Since(now)
	fields := make([]log.Field, 0, 11)
	fields = append(
		fields,
		log.String("host", c.client.HostURL),
		log.String("method", request.Method),
		log.String("uri", uri),
		log.Any("headers", request.Header),
		log.Any("query", request.QueryParam),
		log.Any("body", request.Body),
		log.Float64("quota", quota),
		log.Float64("duration", duration.Seconds()),
		log.Any("reply", response),
		log.Int("code", estatus.Code()),
		log.String("error", estatus.Message()),
	)

	if duration >= c.config.SlowRequestDuration {
		log.Warn(ctx, "http-slow-request-log", fields...)
	} else {
		log.Info(ctx, "http-request-log", fields...)
	}

	return err
}

func (c *Client) Get(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodGet, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

func (c *Client) Post(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodPost, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

func (c *Client) Put(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodPut, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

func (c *Client) Delete(ctx context.Context, req *Request, reply interface{}) error {
	request, err := c.NewRequest(ctx, http.MethodDelete, req, reply)
	if err != nil {
		return err
	}

	return c.execute(ctx, request)
}

// sign algorithm:md5(query params + body + secret + timestamp + nonce)
// Notice:stuff body only when HTTP METHOD is not GET.
// Encode query params to `"bar=baz&foo=quux"` sorted by key in any case
func (c *Client) sign(ctx context.Context, method string, timestamp string, nonce string, req *Request) (string, error) {
	sb := strings.Builder{}
	query := req.QueryParams.Encode()
	bodyStr := ""
	if method != http.MethodGet {
		jsonReq, err := json.Marshal(req.Body)
		if err != nil {
			return "", err
		}
		bodyStr = xstring.BytesToString(jsonReq)
	}

	sb.WriteString(query)
	sb.WriteString(bodyStr)
	sb.WriteString(c.config.Secret)
	sb.WriteString(timestamp)
	sb.WriteString(nonce)
	signStr := sb.String()

	digest := md5.Sum(xstring.StringToBytes(signStr))
	sign := hex.EncodeToString(digest[:])
	log.Info(ctx, "signature info", log.String("sign_str", signStr), log.String("sign", sign))

	return sign, nil
}
