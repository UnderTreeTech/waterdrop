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

	"google.golang.org/grpc/peer"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"google.golang.org/grpc"
)

func (s *Server) Metric() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		var ip string
		if peer, ok := peer.FromContext(ctx); ok {
			ip = peer.Addr.String()
		}

		// call server interceptor
		resp, err = handler(ctx, req)
		estatus := status.ExtractStatus(err)

		metric.UnaryServerHandleCounter.Inc(ip, info.FullMethod, estatus.Error())
		metric.UnaryServerReqDuration.Observe(time.Since(now).Seconds(), ip, info.FullMethod)

		return
	}
}
