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

package server

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/go-playground/assert/v2"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/config"
)

var srv = New(config.DefaultServerConfig())

func TestMain(m *testing.M) {
	defer log.New(nil).Sync()

	code := m.Run()
	os.Exit(code)
}

func TestStart(t *testing.T) {
	net := srv.Start()
	assert.Equal(t, "[::]:9000", net.String())
	assert.Equal(t, "tcp", net.Network())
}

func TestStop(t *testing.T) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := srv.Stop(context.Background())
		assert.Equal(t, err, nil)
	}()

	time.Sleep(200 * time.Millisecond)
}
