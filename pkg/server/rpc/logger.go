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

package rpc

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"google.golang.org/grpc/peer"

	"google.golang.org/grpc"
)

func loggerForUnaryServer(config *ServerConfig) grpc.UnaryServerInterceptor {
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

		fields := make([]log.Field, 0, 8)
		fields = append(
			fields,
			log.String("peer_ip", ip),
			log.String("method", info.FullMethod),
			log.Any("req", req),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.Any("reply", resp),
			log.Int("code", estatus.Code()),
			log.String("error", estatus.Message()),
		)

		if duration >= config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-access-log", fields...)
		} else {
			log.Info(ctx, "grpc-access-log", fields...)
		}

		return
	}
}

func loggerForUnaryClient(config *ClientConfig) grpc.UnaryClientInterceptor {
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
		if estatus.Code() != status.ServiceUnavailable.Code() {
			peerIP = peerInfo.Addr.String()
		}
		fields := make([]log.Field, 0, 8)
		fields = append(
			fields,
			log.String("peer_ip", peerIP),
			log.String("method", method),
			log.Any("req", req),
			log.Float64("quota", quota),
			log.Float64("duration", duration.Seconds()),
			log.Any("reply", reply),
			log.Int("code", estatus.Code()),
			log.String("error", estatus.Message()),
		)

		if duration >= config.SlowRequestDuration {
			log.Warn(ctx, "grpc-slow-request-log", fields...)
		} else {
			log.Info(ctx, "grpc-request-log", fields...)
		}

		return
	}
}
