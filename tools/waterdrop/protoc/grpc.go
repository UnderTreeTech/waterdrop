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

//Notice that you must execute `go get github.com/gogo/protobuf` at $GOPATH/src dir first
// cd $GOPATH/src
// go get -u github.com/gogo/protobuf
const (
	_genGoFastAddress = "go get github.com/gogo/protobuf/protoc-gen-gofast"
	//默认proto生成在.proto文件所在目录
	_grpcProtocCmd = `protoc --proto_path=%s:%s:%s --gofast_out=plugins=grpc:.`
)

func generateGRPC(ctx *cli.Context) error {
	if err := installGogoProtoc(); err != nil {
		return err
	}

	if err := doGenerate(ctx, _grpcProtocCmd); err != nil {
		return err
	}

	return nil
}

func installGogoProtoc() error {
	if _, err := exec.LookPath("protoc-gen-gofast"); err != nil {
		if err = utils.ExecuteGoGet(_genGoFastAddress); err != nil {
			return err
		}
	}

	return nil
}
