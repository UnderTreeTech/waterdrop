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
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/examples/proto/user"
	httpClient "github.com/UnderTreeTech/waterdrop/pkg/server/http/client"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/UnderTreeTech/waterdrop/examples/app/internal/ecode"
	"github.com/UnderTreeTech/waterdrop/examples/proto/demo"
)

func (s *Service) SayHello(ctx context.Context, req *demo.HelloReq) (reply *emptypb.Empty, err error) {
	reply = &emptypb.Empty{}
	return reply, nil
}
func (s *Service) SayHelloURL(ctx context.Context, req *demo.HelloReq) (reply *demo.HelloResp, err error) {
	reply = &demo.HelloResp{Content: "Hello " + req.Name}
	time.Sleep(time.Millisecond * 50)

	r := &httpClient.Request{
		URI:       "/api/app/validate/{id}",
		PathParam: map[string]string{"id": "1"},
		//Body:       `{"email":"example@example.com","name":"John&Sun","password":"styd.cn","sex":2,"age":12,"addr":[{"mobile":"上海市徐汇区","address":"<a onblur='alert(secret)' href='http://www.google.com'>Google</a>","app":{"sappkey":"<p>md5hash</p>"},"reply":{"urls":["www.&baidu.com","www.g&oogle.com","&#x6a;&#x61;&#x76;&#x61;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3a;&#x61;&#x6c;&#x65;&#x72;&#x74;&#x28;&#x31;&#x29;&#x3b;","u003cimg src=1 onerror=alert(/xss/)u003e"]},"resp":[{"app_key":"sha1hash","app_secret":"<href>rsa</href>"}]}]}`,
		Body: &user.ValidateReq{
			Email:    "example@example.com",
			Name:     "John&Sun",
			Password: "styd.cn",
			Sex:      2,
			Age:      12,
			Addr: []*user.Address{
				{
					Address: "<a onblur='alert(secret)' href='http://www.google.com'>Google</a>",
					Mobile:  "上海市徐汇区",
					App: &user.AppReq{
						Sappkey: "<p>md5hash</p>",
					},
					Reply: &user.SkipUrlsReply{
						Urls: []string{"www.&baidu.com",
							"www.g&oogle.com",
							"&#x6a;&#x61;&#x76;&#x61;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3a;&#x61;&#x6c;&#x65;&#x72;&#x74;&#x28;&#x31;&#x29;&#x3b;",
							"u003cimg src=1 onerror=alert(/xss/)u003e",
						},
					},
					Resp: []*user.AppReply{
						{
							Appkey:    "sha1hash",
							Appsecret: "<href>rsa</href>",
						},
					},
				},
			},
		},
	}

	var result interface{}
	err = s.http.Post(ctx, r, &result)
	if err != nil {
		fmt.Println("http response", err)
	}

	return reply, ecode.AppKeyInvalid
}
