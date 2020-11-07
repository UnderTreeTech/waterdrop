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

package interceptors

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type Mock struct {
	Name  string `validate:"required,min=6,max=10"`
	Email string `validate:"required,email"`
	Age   int32  `validate:"required,gte=1,lte=60"`
}

func TestValidate(t *testing.T) {
	interceptor := Validate()
	handler := func(ctx context.Context, req interface{}) (resp interface{}, err error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.testing.TestService/UnaryCall",
	}

	successMock := &Mock{
		Name:  "waterdrop",
		Email: "example@example.com",
		Age:   30,
	}
	t.Run("validate success", func(t *testing.T) {
		resp, err := interceptor(context.Background(), successMock, info, handler)
		assert.Nil(t, resp)
		assert.Nil(t, err)
	})

	failMock := &Mock{
		Name:  "hello",
		Email: "waterdrop",
		Age:   100,
	}
	t.Run("validate fail", func(t *testing.T) {
		resp, err := interceptor(context.Background(), failMock, info, handler)
		assert.Nil(t, resp)
		assert.NotEqual(t, nil, err)
	})
}

func TestGetValidator(t *testing.T) {
	v := GetValidator()
	assert.IsType(t, &validator.Validate{}, v)
}
