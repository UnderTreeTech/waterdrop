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
	"github.com/urfave/cli/v2"
)

// run run commands
func run(ctx *cli.Context) (err error) {
	if err = checkProtocEnv(); err != nil {
		return
	}

	// 根据指定目录下的proto 文件 生成pb.go 文件
	if genGRPC {
		if err = generateGRPC(ctx); err != nil {
			return
		}
	}

	// 根据指定目录下的proto 文件 生成pb.swagger.json 文件
	if genSwagger {
		if err = generateSwagger(ctx); err != nil {
			return
		}
	}

	return
}
