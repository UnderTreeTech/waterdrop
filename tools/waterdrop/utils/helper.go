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

package utils

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func ExecuteGoGet(address string) error {
	args := strings.Split(address, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunTool(name string, dir string, args []string) (err error) {
	cmd := toolPath(name)
	fmt.Println(name, dir, cmd, args)
	toolCmd := &exec.Cmd{
		Path:   cmd,
		Args:   append([]string{cmd}, args...),
		Dir:    dir,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    os.Environ(),
	}
	if filepath.Base(cmd) == cmd {
		var lp string
		if lp, err = exec.LookPath(cmd); err == nil {
			toolCmd.Path = lp
		}
	}
	if err = toolCmd.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); !ok || !e.Exited() {
			fmt.Fprintf(os.Stderr, "运行 %s 出错: %v\n", name, err)
		}
	}
	return
}

func Gopath() (gp string) {
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

func toolPath(toolName string) string {
	gobin := Getenv("GOBIN")
	if runtime.GOOS == "windows" {
		toolName += ".exe"
	}
	if gobin != "" {
		return filepath.Join(gobin, toolName)
	}
	return filepath.Join(Gopath(), "bin", toolName)
}
