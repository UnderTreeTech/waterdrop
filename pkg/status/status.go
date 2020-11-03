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

package status

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	gstatus "google.golang.org/grpc/status"

	// nolint:staticcheck
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

var (
	_status sync.Map

	// predefined global status, it can be override.
	OK                 = add(0, "ok")
	RequestErr         = add(400, "错误请求")
	Unauthorized       = add(401, "未认证，请先登录")
	AccessDenied       = add(403, "未授权访问")
	NothingFound       = add(404, "页面不存在")
	MethodNotAllowed   = add(405, "不支持该方法")
	LimitExceed        = add(429, "请勿频繁请求")
	Canceled           = add(498, "客户端取消请求")
	ServerErr          = add(500, "网络错误，请稍后重试")
	ServiceUnavailable = add(503, "过载保护,服务暂不可用")
	Deadline           = add(504, "服务调用超时")
	AppKeyInvalid      = add(600, "应用程序不存在或已被封禁")
	SignCheckErr       = add(601, "签名校验失败")
	RepeatedRequest    = add(602, "重复请求")
	CaptchaErr         = add(603, "验证码错误")
	TargetBlocked      = add(604, "资源锁定中，请稍后重试")
	PayloadTooLarge    = add(605, "请求体大小超出限制")
	ServiceUpdate      = add(606, "系统升级中")
	UndefinedErr       = add(1000, "未知错误")
)

func New(code int, msg string) *Status {
	if code < 0 {
		panic(fmt.Sprintf("status code must be greater than zero"))
	}

	estatus := add(code, msg)

	return estatus
}

func add(code int, msg string) *Status {
	estatus := new(code, msg)
	_status.Store(code, estatus)

	return estatus
}

// Status represents an RPC status code, message, and details.  It is immutable
// and should be created with New, Newf, or FromProto.
type Status struct {
	s *spb.Status
}

// New returns a Status representing c and msg.
func new(c int, msg string) *Status {
	return &Status{s: &spb.Status{Code: int32(c), Message: msg, Details: make([]*any.Any, 0)}}
}

// FromProto returns a Status representing s.
func FromProto(s *spb.Status) *Status {
	return &Status{s: proto.Clone(s).(*spb.Status)}
}

//implement error interface, return err code
func (s *Status) Error() string {
	return strconv.Itoa(s.Code())
}

// Code returns the status code contained in s.
func (s *Status) Code() int {
	if s == nil || s.s == nil {
		return int(OK.s.Code)
	}

	return int(s.s.Code)
}

// Message returns the message contained in s.
func (s *Status) Message() string {
	if s == nil || s.s == nil {
		return ""
	}

	return s.s.Message
}

// Proto returns s's status as an spb.Status proto message.
func (s *Status) Proto() *spb.Status {
	if s == nil {
		return nil
	}

	return proto.Clone(s.s).(*spb.Status)
}

// WithDetails returns a new status with the provided details messages appended to the status.
// If any errors are encountered, it returns nil and the first error encountered.
func (s *Status) WithDetails(details ...proto.Message) (*Status, error) {
	if s.Code() == OK.Code() {
		return nil, errors.New("no error details for status with code OK")
	}
	// s.Code() != OK implies that s.Proto() != nil.
	p := s.Proto()
	for _, detail := range details {
		any, err := ptypes.MarshalAny(detail)
		if err != nil {
			return nil, err
		}
		p.Details = append(p.Details, any)
	}

	return &Status{s: p}, nil
}

// Details returns a slice of details messages attached to the status.
// If a detail cannot be decoded, the error is returned in place of the detail.
func (s *Status) Details() []interface{} {
	if s == nil || s.s == nil {
		return nil
	}
	details := make([]interface{}, 0, len(s.s.Details))
	for _, any := range s.s.Details {
		detail := &ptypes.DynamicAny{}
		if err := ptypes.UnmarshalAny(any, detail); err != nil {
			details = append(details, err)
			continue
		}
		details = append(details, detail.Message)
	}

	return details
}

// err convert grpc unknown code to ecode status
func errToStatus(code string) *Status {
	ecode, err := strconv.Atoi(code)
	if err != nil {
		log.Errorf("internal_error", log.String("error", code))
		return ServerErr
	}

	estatus, ok := _status.Load(ecode)
	if !ok {
		return UndefinedErr
	}

	return estatus.(*Status)
}

// extract status from grpc call reply err
func ExtractStatus(err error) *Status {
	if err == nil {
		return OK
	}

	gst, _ := gstatus.FromError(err)
	switch gst.Code() {
	case codes.OK:
		return OK
	case codes.InvalidArgument:
		return RequestErr
	case codes.NotFound:
		return NothingFound
	case codes.PermissionDenied:
		return AccessDenied
	case codes.Unauthenticated:
		return Unauthorized
	case codes.ResourceExhausted:
		return LimitExceed
	case codes.Unimplemented:
		return MethodNotAllowed
	case codes.Canceled:
		return Canceled
	case codes.DeadlineExceeded:
		return Deadline
	case codes.Unavailable:
		return ServiceUnavailable
	case codes.Internal:
		return ServerErr
	case codes.Unknown:
		return errToStatus(gst.Message())
	}

	return ServerErr
}

// ExtractContextStatus converts a context error into a Status.  It returns a
// Status with status.OK if err is nil, or a Status from errToStatus if err is
// non-nil and not a context error.
func ExtractContextStatus(err error) *Status {
	switch err {
	case nil:
		return OK
	case context.DeadlineExceeded:
		return Deadline
	case context.Canceled:
		return Canceled
	default:
		return errToStatus(err.Error())
	}
}
