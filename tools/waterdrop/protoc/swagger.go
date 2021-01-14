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
	"os/exec"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/utils"

	"github.com/urfave/cli/v2"
)

const (
	_getSwaggerGen = "go get github.com/UnderTreeTech/waterdrop/tools/waterdrop/protoc/protobuf/protoc-gen-swagger"
	//默认proto生成在.proto文件所在目录
	_swaggerProtoc = `protoc --proto_path=%s:%s:%s --swagger_out=:.`
)

func installSwaggerProtoc() error {
	if _, err := exec.LookPath("protoc-gen-swagger"); err != nil {
		if err := utils.ExecuteGoGet(_getSwaggerGen); err != nil {
			return err
		}
	}
	return nil
}

func generateSwagger(ctx *cli.Context) error {
	if err := installSwaggerProtoc(); err != nil {
		return err
	}

	if err := doGenerate(ctx, _swaggerProtoc); err != nil {
		return err
	}

	return nil
}
