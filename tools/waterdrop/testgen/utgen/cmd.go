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

package utgen

import (
	"fmt"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/testgen/common"

	"github.com/urfave/cli/v2"
)

var genFunc string

var UTCmd = &cli.Command{
	Name:            "utgen",
	Usage:           "waterdrop unit test tools",
	Action:          run,
	SkipFlagParsing: false,
	UsageText:       "ut",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "func",
			Usage:       "whether to generate unit test by func",
			Destination: &genFunc,
		},
	},
}

func run(ctx *cli.Context) error {
	var (
		err    error
		files  []string
		parses []*common.Parse
	)

	if err = common.ParseArgs(ctx.Args().Slice(), &files, 0); err != nil {
		panic(err)
	}

	if parses, err = common.ParseFile(files...); err != nil {
		panic(err)
	}

	if err = genTest(parses); err != nil {
		panic(err)
	}

	fmt.Println(common.GenTestSuccess)
	return nil
}
