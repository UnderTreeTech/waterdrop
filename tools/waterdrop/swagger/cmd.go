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

package swagger

import (
	"os"
	"os/exec"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/utils"

	"github.com/urfave/cli/v2"
)

var _installSwagger = `go get -u github.com/go-swagger/go-swagger/cmd/swagger`

var SwaggerCmd = &cli.Command{
	Name:            "swagger",
	Usage:           "waterdrop swagger tools",
	Action:          run,
	SkipFlagParsing: false,
	UsageText:       "swagger",
}

// run execute swagger serve command
func run(ctx *cli.Context) error {
	if _, err := exec.LookPath("swagger"); err != nil {
		if err = utils.ExecuteGoGet(_installSwagger); err != nil {
			return err
		}
	}

	pwd, _ := os.Getwd()
	return utils.RunTool(ctx, pwd, ctx.Args().Slice())
}
