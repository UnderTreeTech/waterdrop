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

package project

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fatih/color"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/project/base"

	"github.com/urfave/cli/v2"
)

var ProjectCmd = &cli.Command{
	Name:            "new",
	Usage:           "waterdrop project layout tools",
	Action:          run,
	SkipFlagParsing: false,
	UsageText:       "new",
}

func run(c *cli.Context) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if c.Args().Len() == 0 {
		fmt.Println("You must assign project name")
		return
	}

	projectName := c.Args().Slice()[0]
	to := path.Join(pwd, projectName)
	if _, err = os.Stat(to); !os.IsNotExist(err) {
		fmt.Println(fmt.Sprintf("Project %s already exists, please reassign a new name.", color.RedString(projectName)))
		return nil
	}

	repo := base.NewRepo("https://github.com/UnderTreeTech/layout.git", "master")
	if err = repo.CopyTo(ctx, to, path.Base(projectName), []string{".git", ".github"}); err != nil {
		fmt.Println("New project fail, please try again later.")
		return
	}
	base.Tree(to, pwd)

	fmt.Println("\n🍺 🍺 🍺 Project creation succeeded", color.RedString(projectName), "🍺 🍺 🍺 ")
	fmt.Print("💻 Use the following command to start the project 👇:\n\n")
	fmt.Println(color.WhiteString("$ cd %s", projectName))
	fmt.Println(color.WhiteString("$ go mod tidy"))
	fmt.Println(color.WhiteString("$ cd cmd"))
	fmt.Println(color.WhiteString("$ go build -o %s main.go", projectName))
	fmt.Println(color.WhiteString("$ ./%s -conf=../configs/application.toml\n", projectName))
	fmt.Println("🤝 🤝 🤝 Thanks for using Waterdrop 🤝 🤝 🤝 ")
	return
}
