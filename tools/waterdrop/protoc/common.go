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
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// download protoc: https://github.com/protocolbuffers/protobuf/releases.
// copy protoc and include directory to GOPATH/bin
// go get -u github.com/golang/protobuf/protoc-gen-go
func checkProtocEnv() (err error) {
	if _, err = exec.LookPath("protoc"); err != nil {
		err = errors.New("You haven't installed Protobuf yet，Please visit this page to install with your own system：https://github.com/protocolbuffers/protobuf/releases")
		return err
	}
	return nil
}

func doGenerate(ctx *cli.Context, protocCmd string) (err error) {
	files := ctx.Args().Slice()
	if len(files) == 0 {
		files, _ = filepath.Glob("*.proto")
	}

	pwd, _ := os.Getwd()
	// case go path
	gosrc := filepath.Join(build.Default.GOPATH, "src")
	_, err = os.Stat(gosrc)
	if err != nil {
		fmt.Println("src directory does not exist, please create it in your GOPATH")
		return nil
	}

	cmdLine := fmt.Sprintf(protocCmd, gosrc, pwd)
	args := strings.Split(cmdLine, " ")
	args = append(args, files...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = pwd
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return
}
