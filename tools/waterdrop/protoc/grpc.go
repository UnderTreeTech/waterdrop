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

package protoc

import (
	"github.com/urfave/cli/v2"
)

const (
	//default generate pb files to current directory
	_grpcProtocCmd = `protoc -I=%s -I=%s --go_out=plugins=grpc:.`
)

// generateGRPC generate grpc pb files
func generateGRPC(ctx *cli.Context) error {
	if err := doGenerate(ctx, _grpcProtocCmd); err != nil {
		return err
	}

	return nil
}
