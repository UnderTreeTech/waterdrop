package xreply

import (
	"github.com/UnderTreeTech/waterdrop/pkg/status"
	"github.com/golang/protobuf/ptypes/empty"
)

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var emptyReply = &empty.Empty{}

func Reply(data interface{}, err error) interface{} {
	if err != nil {
		return exception(err)
	}

	return success(data)
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
