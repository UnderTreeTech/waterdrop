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

package service

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/ecode"

	"github.com/UnderTreeTech/protobuf/demo"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) SayHello(ctx context.Context, req *demo.HelloReq) (reply *empty.Empty, err error) {
	reply = &empty.Empty{}
	return reply, nil
}
func (s *Service) SayHelloURL(ctx context.Context, req *demo.HelloReq) (reply *demo.HelloResp, err error) {
	reply = &demo.HelloResp{Content: "Hello " + req.Name}
	time.Sleep(time.Millisecond * 999)
	return reply, ecode.AppKeyInvalid
}
