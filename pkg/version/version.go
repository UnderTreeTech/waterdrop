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

package version

import (
	"fmt"
	"os"
	"runtime"
)

const (
	// Version is waterdrop current version
	Version = "v1.3.6"
	// unknownProperty
	unknownProperty = ""
	// Compiler is a convenient alias for runtime.Compiler.
	Compiler = runtime.Compiler
)

var (
	// Name is meant to be injected with the binary's intended name, by means
	// of `go -ldflags` at build-time
	Name = unknownProperty
	// BuildBranch is meant to be injected with a string denoting the build branch
	// of the binary, by means of `go -ldflags` at build-time
	BuildBranch = unknownProperty
	// GoVersion is the version of the Go toolchain used to build the binary
	// (e.g. "go1.19.2")
	// It defaults to value of runtime.Version() if not explicitly overridden
	GoVersion = unknownProperty
	// GitCommit is the commit hash of the Git repository's HEAD at build-time
	GitCommit = unknownProperty
	// SDKGitCommit is the commit hash of the sdk Git repository's HEAD at build-time
	SDKGitCommit = unknownProperty
	// BuildDate is meant to be injected with a string denoting the build time
	// of the binary, by means of `go -ldflags` at build-time
	BuildDate = unknownProperty
	// BuildComments can be used to associate arbitrary extra information with
	// the binary, by means of injection via `go -ldflags` at build-time
	BuildComments = unknownProperty
	// Platform is a string in the form of "GOOS/GOARCH", e.g. "linux/amd64"
	Platform = unknownProperty
)

// This is for preventing access to the unpopulated properties
func init() {
	// usages: <command> version (or --version / -version)
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-version") {
		printBuildInfo()
		os.Exit(0)
	}
}

// printBuildInfo prints out the collected version information
func printBuildInfo() {
	printf := func(k string, v string) {
		fmt.Printf("%s:\t%s\n", k, v)
	}

	if Name != unknownProperty {
		printf("Name", Name)
	}

	if BuildBranch != unknownProperty {
		printf("Build branch", BuildBranch)
	}

	if GoVersion == unknownProperty {
		GoVersion = runtime.Version()
	}
	printf("Go version", GoVersion)

	if GitCommit != unknownProperty {
		printf("Git commit", BuildBranch)
	}

	if SDKGitCommit != unknownProperty {
		printf("SDK commit", BuildBranch)
	}

	if BuildDate != unknownProperty {
		printf("Build date", BuildDate)
	}

	if BuildComments != unknownProperty {
		printf("Build comments", BuildComments)
	}

	if Platform == unknownProperty {
		Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}
	printf("OS/Arch", Platform)
	printf("Compiler", Compiler)
}
