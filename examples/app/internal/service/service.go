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
	"fmt"

	"github.com/UnderTreeTech/waterdrop/examples/proto/user"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/server/http/client"
	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
	rpcClient "github.com/UnderTreeTech/waterdrop/pkg/server/rpc/client"
	rpcConfig "github.com/UnderTreeTech/waterdrop/pkg/server/rpc/config"
)

type Service struct {
	user user.UserClient
	http *client.Client
}

func New() *Service {
	cliConf := &rpcConfig.ClientConfig{}
	if err := conf.Unmarshal("client.rpc.user", cliConf); err != nil {
		panic(fmt.Sprintf("unmarshal user client config fail, err msg %s", err.Error()))
	}
	rpcCli := rpcClient.New(cliConf)
	userRPC := user.NewUserClient(rpcCli.GetConn())

	httpCliConf := &config.ClientConfig{}
	if err := conf.Unmarshal("client.http.app", httpCliConf); err != nil {
		panic(fmt.Sprintf("unmarshal http client config fail, err msg %s", err.Error()))
	}
	httpCli := client.New(httpCliConf)

	return &Service{
		http: httpCli,
		user: userRPC,
	}
}
