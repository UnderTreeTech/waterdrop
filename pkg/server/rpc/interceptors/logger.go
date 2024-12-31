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

package interceptors

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xslice"

	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"google.golang.org/grpc/peer"

	"google.golang.org/grpc"
)

// LoggerForUnaryServer log unary server response details
func LoggerForUnaryServer(config *config.ServerConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		var ip string
		if peer, ok := peer.FromContext(ctx); ok {
			ip = peer.Addr.String()
		}

		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		// call server interceptor
		resp, err = handler(ctx, req)

		estatus := status.ExtractStatus(err)
		duration := time.Since(now)

		method := info.FullMethod
		fields := make([]log.Field, 0, 8)
		fields = append(
			fields,
			log.String("peer", ip),
			log.String("method", method),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.Any("reply", json.RawMessage(log.JsonBytes(resp))),
			log.Int("code", estatus.Code()),
			log.String("error", estatus.Message()),
		)

		details := strings.Split(method, "/")
		fnName := details[len(details)-1]
		if !xslice.ContainString(config.NotLog, fnName) {
			fields = append(fields, log.Any("req", json.RawMessage(log.JsonBytes(req))))
		}

		if duration >= config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-access-log", fields...)
		} else {
			log.Info(ctx, "grpc-access-log", fields...)
		}
		return
	}
}

// LoggerForUnaryClient log unary client request details
func LoggerForUnaryClient(config *config.ClientConfig) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		now := time.Now()

		var peerInfo peer.Peer
		opts = append(opts, grpc.Peer(&peerInfo))

		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		// call client interceptor
		err = invoker(ctx, method, req, reply, cc, opts...)

		estatus := status.ExtractStatus(err)
		duration := time.Since(now)
		var peerIP string
		if peerInfo.Addr != nil {
			peerIP = peerInfo.Addr.String()
		}
		fields := make([]log.Field, 0, 7)
		fields = append(
			fields,
			log.String("peer", peerIP),
			log.String("method", method),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.Int("code", estatus.Code()),
			log.String("error", estatus.Message()),
		)

		details := strings.Split(method, "/")
		fnName := details[len(details)-1]
		if !xslice.ContainString(config.NotLog, fnName) {
			fields = append(fields, log.Any("req", json.RawMessage(log.JsonBytes(req))))
		}

		if duration >= config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-request-log", fields...)
		} else {
			log.Info(ctx, "grpc-request-log", fields...)
		}
		return
	}
}
