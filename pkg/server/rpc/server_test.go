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

package rpc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"google.golang.org/grpc/stats/grpc_testing"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

var srv = NewServer(defaultServerConfig())

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()

	code := m.Run()
	os.Exit(code)
}

func TestStart(t *testing.T) {
	grpc_testing.RegisterTestServiceServer(srv.server, &grpc_testing.UnimplementedTestServiceServer{})
	net := srv.Start()
	assert.Equal(t, "[::]:20812", net.String())
	assert.Equal(t, "tcp", net.Network())
}

func TestStop(t *testing.T) {
	go func() {
		time.Sleep(1 * time.Second)
		err := srv.Stop(context.Background())
		assert.Equal(t, err, nil)
	}()

	time.Sleep(200 * time.Millisecond)
}
