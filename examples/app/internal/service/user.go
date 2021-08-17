/*
 *
 * Copyright 2021 waterdrop authors.
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

	"github.com/UnderTreeTech/waterdrop/examples/proto/user"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetStaffInfo(ctx context.Context, req *user.StaffInfoReq) (reply *user.StaffInfoReply, err error) {
	return
}

func (s *Service) DelStaff(ctx context.Context, req *user.StaffInfoReq) (reply *emptypb.Empty, err error) {
	reply = &emptypb.Empty{}
	if _, err = s.user.TestValidator(ctx, &user.ValidateReq{}); err != nil {
		fmt.Println("test validator", err)
	}
	return
}

func (s *Service) GetAppSecret(ctx context.Context, req *user.AppReq) (reply *user.AppReply, err error) {
	return
}

func (s *Service) GetAppSkipUrls(ctx context.Context, req *user.AppReq) (reply *user.SkipUrlsReply, err error) {
	return
}

func (s *Service) TestValidator(ctx context.Context, req *user.ValidateReq) (reply *emptypb.Empty, err error) {
	reply = &emptypb.Empty{}
	return
}
