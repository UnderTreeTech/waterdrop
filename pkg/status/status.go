package status

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	gstatus "google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/any"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
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
	Canceled           = add(498, "客户端取消请求")
	ServerErr          = add(500, "网络错误，请稍后重试")
	ServiceUnavailable = add(503, "过载保护,服务暂不可用")
	Deadline           = add(504, "服务调用超时")
	LimitExceed        = add(509, "请求超出限制")
	UndefinedErr       = add(600, "未知错误")
)

func New(code int, msg string) *status {
	if code < 0 {
		panic(fmt.Sprintf("status code must be greater than zero"))
	}

	estatus := add(code, msg)

	return estatus
}

func add(code int, msg string) *status {
	estatus := new(code, msg)
	_status.Store(code, estatus)

	return estatus
}

// Status represents an RPC status code, message, and details.  It is immutable
// and should be created with New, Newf, or FromProto.
type status struct {
	s *spb.Status
}

// New returns a Status representing c and msg.
func new(c int, msg string) *status {
	return &status{s: &spb.Status{Code: int32(c), Message: msg, Details: make([]*any.Any, 0)}}
}

// FromProto returns a Status representing s.
func FromProto(s *spb.Status) *status {
	return &status{s: proto.Clone(s).(*spb.Status)}
}

//implement error interface, return err code
func (s *status) Error() string {
	return strconv.Itoa(s.Code())
}

// Code returns the status code contained in s.
func (s *status) Code() int {
	if s == nil || s.s == nil {
		return int(OK.s.Code)
	}

	return int(s.s.Code)
}

// Message returns the message contained in s.
func (s *status) Message() string {
	if s == nil || s.s == nil {
		return ""
	}

	return s.s.Message
}

// Proto returns s's status as an spb.Status proto message.
func (s *status) Proto() *spb.Status {
	if s == nil {
		return nil
	}

	return proto.Clone(s.s).(*spb.Status)
}

// WithDetails returns a new status with the provided details messages appended to the status.
// If any errors are encountered, it returns nil and the first error encountered.
func (s *status) WithDetails(details ...proto.Message) (*status, error) {
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

	return &status{s: p}, nil
}

// Details returns a slice of details messages attached to the status.
// If a detail cannot be decoded, the error is returned in place of the detail.
func (s *status) Details() []interface{} {
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

// err convert grpc unkown code to ecode status
func errToStatus(code string) *status {
	ecode, err := strconv.Atoi(code)
	if err != nil {
		log.Errorf("internal_error", log.String("error", code))
		return ServerErr
	}

	estatus, ok := _status.Load(ecode)
	if !ok {
		return UndefinedErr
	}

	return estatus.(*status)
}

// extract status from grpc call reply err
func ExtractStatus(err error) *status {
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
	case codes.Unknown:
		return errToStatus(gst.Message())
	}

	return ServerErr
}
