package harness

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/project/base"
)

const defaultEcodeCSVContent = `0,SUCCESS,成功,zh-cn
0,SUCCESS,"Success",en
100000,InternalError,服务器开了小差，请稍后重试,zh-cn
100000,InternalError,"Network is unavailable, please have a try later",en
100001,InvalidParam,参数错误,zh-cn
100001,InvalidParam,"Invalid parameters, please check your request",en
`

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

	ecodeDir := path.Join(to, "api", "ecode")

	// Create harness directories for api/idl and api/ecode
	if err := os.MkdirAll(path.Join(to, "api", "idl"), 0755); err != nil {
		fmt.Println("Create harness directories failed:", err)
		return err
	}
	if err := os.MkdirAll(ecodeDir, 0755); err != nil {
		fmt.Println("Create harness ecode directory failed:", err)
		return err
	}

	// Create go.work
	goWorkContent := `go 1.25.0

use (
	./api/idl
	./api/ecode
)
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

	// Create api/ecode/go.mod
	ecodeGoModContent := fmt.Sprintf("module %s\n\ngo 1.25.0\n", "ecode")
	if err := os.WriteFile(path.Join(ecodeDir, "go.mod"), []byte(ecodeGoModContent), 0644); err != nil {
		fmt.Println("Create api/ecode/go.mod failed:", err)
		return err
	}

	ecodeCSVPath := path.Join(ecodeDir, "ecode.csv")
	if err := os.WriteFile(ecodeCSVPath, []byte(defaultEcodeCSVContent), 0644); err != nil {
		fmt.Println("Create api/ecode/ecode.csv failed:", err)
		return err
	}

	// Create api/ecode/README.md
	ecodeReadmeContent := fmt.Sprintf("# %s Ecode\n\nThis directory contains business error code definitions for the `%s` harness.\n\nEdit `ecode.csv` and run `waterdrop ecode ./ecode.csv .` in this directory to regenerate `ecode.go`.\n\nThe initial `ecode.csv` is embedded in the harness generator.\n", harnessName, harnessName)
	if err := os.WriteFile(path.Join(ecodeDir, "README.md"), []byte(ecodeReadmeContent), 0644); err != nil {
		fmt.Println("Create api/ecode/README.md failed:", err)
		return err
	}

	if err := generateEcode(ecodeDir, ecodeCSVPath); err != nil {
		fmt.Println("Generate harness ecode failed:", err)
		return err
	}

	// Show tree
	base.Tree(to, pwd)

	// Print success messages
	fmt.Println("\n🍺 🍺 🍺 Harness creation succeeded", color.RedString(harnessName), "🍺 🍺 🍺 ")
	fmt.Print("💻 Use the following commands to start developing 👇:\n\n")
	fmt.Println(color.WhiteString("$ cd %s", harnessName))
	fmt.Println(color.WhiteString("$ waterdrop new {service_name}"))
	fmt.Println(color.WhiteString("$ go work use ./{service_name}"))
	fmt.Println(color.WhiteString("$ cd ./api/ecode && waterdrop ecode ./ecode.csv .\n"))
	fmt.Println("🤝 🤝 🤝 Thanks for using Waterdrop 🤝 🤝 🤝 ")

	return nil
}

func generateEcode(outputDir, csvPath string) error {
	cmd := exec.Command("waterdrop", "ecode", csvPath, outputDir)
	cmd.Dir = outputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
