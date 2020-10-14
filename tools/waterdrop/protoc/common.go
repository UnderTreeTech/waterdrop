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

func checkProtocEnv() (err error) {
	if _, err = exec.LookPath("protoc"); err != nil {
		err = errors.New("You haven't installed Protobuf yet，Please visit this page to install with your own system：https://github.com/protocolbuffers/protobuf/releases")
		return err
	}
	return nil
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

func doGenerate(ctx *cli.Context, protocCmd string) (err error) {
	files := ctx.Args().Slice()
	if len(files) == 0 {
		files, _ = filepath.Glob("*.proto")
	}

	pwd, _ := os.Getwd()
	gosrc := path.Join(gopath(), "src")
	ext := path.Join(gopath(), "pkg/mod")
	cmdLine := fmt.Sprintf(protocCmd, pwd, gosrc, ext)

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
