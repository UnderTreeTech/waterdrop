package service

import (
	"context"

	"github.com/UnderTreeTech/protobuf/demo"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) SayHello(ctx context.Context, req *demo.HelloReq) (reply *empty.Empty, err error) {
	reply = &empty.Empty{}
	return reply, nil
}
func (s *Service) SayHelloURL(ctx context.Context, req *demo.HelloReq) (reply *demo.HelloResp, err error) {
	reply = &demo.HelloResp{}
	return reply, nil
}
