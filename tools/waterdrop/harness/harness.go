package harness

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/project/base"
)

var HarnessCmd = &cli.Command{
	Name:            "harness",
	Usage:           "waterdrop harness tools for multi-service workspaces",
	SkipFlagParsing: false,
	Subcommands: []*cli.Command{
		{
			Name:   "new",
			Usage:  "create a new harness skeleton",
			Action: runNewHarness,
		},
	},
}

func runNewHarness(c *cli.Context) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if c.Args().Len() == 0 {
		fmt.Println("You must assign harness name")
		return
	}

	harnessName := c.Args().Slice()[0]
	to := path.Join(pwd, harnessName)
	if _, err = os.Stat(to); !os.IsNotExist(err) {
		fmt.Println(fmt.Sprintf("Harness %s already exists, please reassign a new name.", color.RedString(harnessName)))
		return nil
	}

	repo := base.NewRepo("https://github.com/UnderTreeTech/layout.git", "master")

	// Copy harness template
	replaces := []string{
		"monorepo-layout", harnessName,
	}
	ignores := []string{".git", ".github"}

	if err = repo.CopySubDirTo(ctx, ".harness_template", to, replaces, ignores); err != nil {
		fmt.Println("New harness template copy fail, please try again later.", err)
		return err
	}

	// Create harness directories for api/idl
	if err := os.MkdirAll(path.Join(to, "api", "idl"), 0755); err != nil {
		fmt.Println("Create harness directories failed:", err)
		return err
	}

	// Create go.work
	goWorkContent := `go 1.25.0

use ./api/idl
`
	if err := os.WriteFile(path.Join(to, "go.work"), []byte(goWorkContent), 0644); err != nil {
		fmt.Println("Create go.work failed:", err)
		return err
	}

	// Create api/idl/go.mod
	idlGoModContent := fmt.Sprintf("module %s\n\ngo 1.25.0\n", "idl")
	if err := os.WriteFile(path.Join(to, "api", "idl", "go.mod"), []byte(idlGoModContent), 0644); err != nil {
		fmt.Println("Create api/idl/go.mod failed:", err)
		return err
	}

	// Create api/idl/README.md
	idlReadmeContent := fmt.Sprintf("# %s API IDL\n\nThis directory contains the protocol buffers (Protobuf) definitions for the `%s` harness.\n\nRun `waterdrop protoc` commands here to generate Go code.\n", harnessName, harnessName)
	if err := os.WriteFile(path.Join(to, "api", "idl", "README.md"), []byte(idlReadmeContent), 0644); err != nil {
		fmt.Println("Create api/idl/README.md failed:", err)
		return err
	}

	// Show tree
	base.Tree(to, pwd)

	// Print success messages
	fmt.Println("\n🍺 🍺 🍺 Harness creation succeeded", color.RedString(harnessName), "🍺 🍺 🍺 ")
	fmt.Print("💻 Use the following commands to start developing 👇:\n\n")
	fmt.Println(color.WhiteString("$ cd %s", harnessName))
	fmt.Println(color.WhiteString("$ waterdrop new {service_name}"))
	fmt.Println(color.WhiteString("$ go work use ./{service_name}\n"))
	fmt.Println("🤝 🤝 🤝 Thanks for using Waterdrop 🤝 🤝 🤝 ")

	return nil
}
