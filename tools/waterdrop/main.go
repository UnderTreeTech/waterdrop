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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/ecode"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/upgrade"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/project"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/swagger"
	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/testgen/utgen"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/protoc"

	"github.com/urfave/cli/v2"
)

const Version = "v0.2.0"

// main tool entry point
func main() {
	app := cli.NewApp()
	app.Name = "waterdrop"
	app.Usage = "waterdrop tools"
	app.Version = Version
	app.Commands = []*cli.Command{
		project.ProjectCmd,
		protoc.ProtocCmd,
		swagger.SwaggerCmd,
		utgen.UTCmd,
		upgrade.UpgradeCmd,
		ecode.EcodeCmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf(fmt.Sprintf("run waterdrop tool fail, error is %s", err.Error()))
	}
}
