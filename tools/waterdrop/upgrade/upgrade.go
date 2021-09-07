/*
 *
 * Copyright 2021 waterdrop authors.
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

package upgrade

import (
	"fmt"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/project/base"
	"github.com/urfave/cli/v2"
)

var UpgradeCmd = &cli.Command{
	Name:            "upgrade",
	Usage:           "waterdrop upgrade tools",
	Action:          run,
	SkipFlagParsing: false,
	UsageText:       "upgrade",
}

func run(c *cli.Context) (err error) {
	err = base.GoInstall(
		"github.com/UnderTreeTech/waterdrop/tools/waterdrop",
	)

	if err != nil {
		fmt.Println("upgrade waterdrop fail", err.Error())
	}
	return
}
