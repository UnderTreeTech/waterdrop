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

package common

import (
	"fmt"
	"strings"

	"golang.org/x/tools/imports"
)

const GenTestSuccess = `
	      江城子·程序记梦
	十年编程两茫茫，工期短，需求长。
	   百行代码，Bug到处藏。
	纵使上线又如何，新版本，继续忙。
	黑白颠倒没商量，趴桌上，进梦乡。
	   夜半梦醒，无人在身旁。
	最怕灯火阑珊时，手机响，心里慌。
// Generation success. Powered by Waterdrop
`

// GoImport Use golang.org/x/tools/imports auto import pkg
func GoImport(file string, bytes []byte) (res []byte, err error) {
	options := &imports.Options{
		TabWidth:  8,
		TabIndent: true,
		Comments:  true,
		Fragment:  true,
	}
	if res, err = imports.Process(file, bytes, options); err != nil {
		fmt.Printf("GoImport(%s) error(%v)", file, err)
		res = bytes
		return
	}
	return
}

// ConvertMethod checkout the file belongs to dao or not
func ConvertMethod(path string) (method string) {
	switch {
	case strings.Contains(path, "/dao"):
		method = "d"
	case strings.Contains(path, "/service"):
		method = "s"
	default:
		method = ""
	}
	return
}
