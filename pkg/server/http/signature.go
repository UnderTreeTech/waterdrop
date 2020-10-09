/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package http

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/database/redis"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xreply"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
	"github.com/gin-gonic/gin"
	rr "github.com/gomodule/redigo/redis"
)

// Signature middleware is commonly used for p2p communication, like ios/android application, or server to server call
func (s *SignVerify) Signature() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.verify(c); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, xreply.Reply(c.Request.Context(), nil, err))
			return
		}

		c.Next()
	}
}

type SignVerify struct {
	lock      sync.RWMutex
	keys      map[string]string
	skips     []string
	r         *redis.Redis
	client    *Client
	secretURL string
	skipsURL  string
	timeout   int
}

func NewSignatureVerify(config *ClientConfig, r *redis.Redis) *SignVerify {
	client := NewClient(config)
	sv := &SignVerify{
		keys: make(map[string]string),
		//keys: map[string]string{
		//	"xHf74ZfV43cAUsUl": "d0dbe915091d400bd8ee7f27f0791303",
		//},
		skips:     make([]string, 0),
		r:         r,
		client:    client,
		secretURL: _secretURL,
		skipsURL:  _skipsURL,
		timeout:   _requestTimeout,
	}
	return sv
}

func (s *SignVerify) verify(c *gin.Context) (err error) {
	ctx := c.Request.Context()
	nonce := c.Request.Header.Get(_nonce)
	if _nonceLen != len(c.Request.Header.Get(_nonce)) {
		log.Warn(ctx, "invalid nonce", log.String("nonce", nonce))
		return status.RequestErr
	}

	timestamp, err := strconv.Atoi(c.Request.Header.Get(_timestamp))
	if err != nil {
		log.Warn(ctx, "invalid timestamp", log.String("timestamp", c.Request.Header.Get("timestamp")))
		return status.RequestErr
	}

	if timestamp < 0 {
		timestamp = -timestamp
	}

	if time.Now().Unix()-int64(timestamp) > int64(s.timeout) {
		log.Warn(ctx, "request timeout", log.Int64("current_time", time.Now().Unix()), log.Int("request_time", timestamp))
		return status.RequestErr
	}

	//check whether it's a reply attack or not
	exist, err := rr.Int(s.r.Do(ctx, "exists", nonce))
	if exist > 0 {
		log.Warn(ctx, "reply attack", log.String("nonce", nonce))
		return status.RepeatedRequest
	} else {
		//set request sign to redis to avoid reply attack
		s.r.Do(ctx, "setex", nonce, s.timeout, "")
	}

	return s.validSign(c)
}

func (s *SignVerify) validSign(c *gin.Context) error {
	ctx := c.Request.Context()
	sign := c.Request.Header.Get(_sign)
	nonce := c.Request.Header.Get(_nonce)
	ts := c.Request.Header.Get(_timestamp)

	secret, err := s.appSecret(c)
	if err != nil {
		return err
	}

	// sign algorithm:md5(query params + body + secret + timestamp + nonce)
	// Notice:stuff body only when HTTP METHOD is not GET.
	sb := strings.Builder{}
	query := c.Request.URL.Query().Encode()
	body := ""
	if http.MethodGet != c.Request.Method {
		if c.Request.ContentLength > _maxBytes {
			return status.PayloadTooLarge
		}

		var bodyBytes []byte
		if c.Request.ContentLength > 0 {
			bodyBytes = make([]byte, c.Request.ContentLength, c.Request.ContentLength)
			_, err = io.ReadFull(c.Request.Body, bodyBytes)
		} else {
			bodyBytes, err = ioutil.ReadAll(c.Request.Body)
		}

		if err != nil {
			return err
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		body = xstring.BytesToString(bodyBytes)
	}

	sb.WriteString(query)
	sb.WriteString(body)
	sb.WriteString(secret)
	sb.WriteString(ts)
	sb.WriteString(nonce)
	signStr := sb.String()

	log.Debug(ctx, "sign content", log.String("str", signStr))
	digest := md5.Sum(xstring.StringToBytes(signStr))
	if hsign := hex.EncodeToString(digest[:]); hsign != sign {
		log.Warn(ctx, "fake request", log.String("sign", sign), log.String("hsign", hsign))
		return status.SignCheckErr
	}

	return nil
}

func (s *SignVerify) appSecret(c *gin.Context) (string, error) {
	appkey := c.Request.Header.Get(_appkey)
	s.lock.RLock()
	secret, ok := s.keys[appkey]
	s.lock.RUnlock()

	//if appkey-appsecret not exist in memory, fetch it from remote service for once.
	//if fetch reply any kind of err, return AppIdInvalid error
	if !ok {
		s.lock.Lock()
		fetchSecret, err := s.fetchAppSecret(c)
		if err != nil {
			log.Error(c.Request.Context(), "fetch appsecret fail", log.String("invalid_appkey", appkey), log.String("error", err.Error()))
			s.lock.Unlock()
			return "", status.AppKeyInvalid
		}
		secret = fetchSecret
		s.keys[appkey] = fetchSecret
		s.lock.Unlock()
	}

	return secret, nil
}

//fetch appsecret from remote service
func (s *SignVerify) fetchAppSecret(c *gin.Context) (string, error) {
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AppKey    string `json:"app_key"`
			AppSecret string `json:"app_secret"`
		} `json:"data"`
	}

	params := url.Values{}
	params.Set("appkey", c.Request.Header.Get(_appkey))
	req := &Request{
		URI: s.secretURL,
	}

	err := s.client.Get(c.Request.Context(), req, &resp)
	if err != nil {
		return "", status.ExtractContextStatus(err)
	}

	if resp.Code != 0 || resp.Data.AppSecret == "" {
		log.Error(c.Request.Context(), "fetch appsecret fail", log.Int("code", resp.Code), log.String("message", resp.Message), log.Any("data", resp.Data))
		return "", status.ServerErr
	}

	return resp.Data.AppSecret, nil
}

func (s *SignVerify) SetSecretURL(url string) {
	s.secretURL = url
}

func (s *SignVerify) SetSkipsURL(url string) {
	s.skipsURL = url
}

func (s *SignVerify) SetRequestTimeout(timeout int) {
	s.timeout = timeout
}
