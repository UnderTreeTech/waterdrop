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

package xreply

import (
	"context"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/golang/protobuf/ptypes/empty"
)

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var emptyReply = &empty.Empty{}

func Reply(ctx context.Context, data interface{}, err error) interface{} {
	var reply *response
	if err != nil {
		reply = exception(err)
	} else {
		reply = success(data)
	}
	log.Debug(ctx, "reply", log.Int("code", reply.Code), log.String("message", reply.Message), log.Any("data", reply.Data))

	return reply
}

func exception(err error) (resp *response) {
	var estatus *status.Status
	var ok bool

	estatus, ok = err.(*status.Status)
	if !ok {
		estatus = status.ServerErr
	}

	resp = &response{
		Code:    estatus.Code(),
		Message: estatus.Message(),
		Data:    emptyReply,
	}
	return
}

func success(data interface{}) (resp *response) {
	resp = &response{
		Code:    status.OK.Code(),
		Message: status.OK.Message(),
		Data:    data,
	}
	return
}
