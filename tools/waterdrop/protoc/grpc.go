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
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

//Notice that you must execute `go get -u github.com/gogo/protobuf` at $GOPATH/src dir first
// cd $GOPATH/src
// go get -u github.com/gogo/protobuf
const (
	_genGoFastAddress = "go get -u github.com/gogo/protobuf/protoc-gen-gofast"
	//默认proto生成在.proto文件所在目录
	_grpcProtocCmd = `protoc --proto_path=%s:%s:%s --gofast_out=plugins=grpc:.`
)

func checkProtocEnv() (err error) {
	if _, err = exec.LookPath("protoc"); err != nil {
		err = errors.New("You haven't installed Protobuf yet，Please visit this page to install with your own system：https://github.com/protocolbuffers/protobuf/releases")
		return err
	}
	return nil
}

func generateGRPC(ctx *cli.Context) error {
	if err := installGogoProtoc(); err != nil {
		return err
	}

	if err := doGenerate(ctx); err != nil {
		return err
	}

	return nil
}

func installGogoProtoc() error {
	if _, err := exec.LookPath("protoc-gen-gofast"); err != nil {
		if err = executeGoGet(_genGoFastAddress); err != nil {
			return err
		}

		if os.Getenv("GO111MODULE") != "off" {
			if moderr := executeGoGet(_genGoFastAddress + "@v1.3.1"); moderr != nil {
				return moderr
			}
		}
	}

	return nil
}

func executeGoGet(address string) error {
	args := strings.Split(address, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func doGenerate(ctx *cli.Context) (err error) {
	files := ctx.Args().Slice()
	if len(files) == 0 {
		files, _ = filepath.Glob("*.proto")
	}

	pwd, _ := os.Getwd()
	gosrc := path.Join(gopath(), "src")
	ext := path.Join(gopath(), "pkg/mod")
	cmdLine := fmt.Sprintf(_grpcProtocCmd, pwd, gosrc, ext)

	args := strings.Split(cmdLine, " ")
	args = append(args, files...)
	cmd := exec.Command(args[0], args[1:]...)
	fmt.Println("cmd", cmd)
	cmd.Dir = pwd
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return
}

func gopath() (gp string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))

	if len(gopaths) == 1 && gopaths[0] != "" {
		return gopaths[0]
	}
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	abspwd, err := filepath.Abs(pwd)
	if err != nil {
		return
	}
	for _, gopath := range gopaths {
		if gopath == "" {
			continue
		}
		absgp, err := filepath.Abs(gopath)
		if err != nil {
			return
		}
		if strings.HasPrefix(abspwd, absgp) {
			return absgp
		}
	}
	return build.Default.GOPATH
}
