/*
 *
 * Copyright 2023 waterdrop authors.
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

package ecode

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/urfave/cli/v2"
)

type Ecode struct {
	ErrName  string
	ErrCode  int
	ErrMsg   string
	Language string
}

var EcodeCmd = &cli.Command{
	Name:            "ecode",
	Usage:           "waterdrop ecode i18n tools",
	Action:          run,
	SkipFlagParsing: false,
	UsageText:       "ecode",
}

// run run commands
func run(c *cli.Context) (err error) {
	if c.Args().Len() == 0 {
		fmt.Println("You must assign ecode file csv path")
		return
	}

	csvPath := c.Args().Slice()[0]
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatal("open file fail", err)
	}
	defer f.Close()

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	packageName := filepath.Base(pwd)

	ecodes := make([]Ecode, 0)
	reader := csv.NewReader(bufio.NewReader(f))
	reader.LazyQuotes = true
	locales := make(map[string]string)
	for {
		details, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal("read line fail", err)
		}
		if len(details) != 4 {
			continue
		}

		code, _ := strconv.Atoi(details[0])
		ecode := Ecode{
			ErrName:  details[1],
			ErrCode:  code,
			ErrMsg:   details[2],
			Language: details[3],
		}
		ecodes = append(ecodes, ecode)

		locales[details[3]] = strings.ToUpper(details[3])
		if strings.Contains(details[3], "-") {
			locales[details[3]] = strings.ToUpper(strings.ReplaceAll(details[3], "-", ""))
		}
	}

	b := &bytes.Buffer{}

	data := struct {
		Package string
		Import  []string
		Ecodes  []Ecode
		Locales map[string]string
		Version int64
	}{
		Package: packageName,
		Import:  []string{"github.com/UnderTreeTech/waterdrop/pkg/status"},
		Ecodes:  ecodes,
		Locales: locales,
		Version: time.Now().Unix(),
	}

	tmpl := `package {{ .Package }}

// Code generated by template. DO NOT EDIT. YOU ONLY NEED EDIT i18n.csv
import (
	{{ range $_, $v := .Import -}}
		"{{ $v }}"
	{{ end -}}
)

const EcodeVersion = {{ .Version }}

type errors map[int]string

var (
	{{- range $_, $v := .Ecodes }}
		{{- if eq $v.Language "zh-cn" }} 
			{{ $v.ErrName }}  = status.New({{ $v.ErrCode }}, "{{ $v.ErrMsg }}") 
		{{- end -}}
	{{ end -}}
)

{{ range $k, $v := .Locales }}
	var {{ $v }}ErrMsg = errors {
		{{- range $_, $e := $.Ecodes }}
			{{- if eq $k $e.Language -}}
				{{ $e.ErrName }}.Code():  "{{ $e.ErrMsg }}",
			{{ end -}}
		{{ end -}}
	}
{{ end }}`

	err = template.Must(template.New("errors").Parse(tmpl)).Execute(b, data)
	if err != nil {
		log.Fatal("generate code fail", err)
	}

	output := path.Join(pwd, "ecode.go")
	if c.Args().Len() == 2 {
		output = path.Join(c.Args().Slice()[1], "ecode.go")
	}

	file, err := os.Create(output)
	if err != nil {
		log.Fatal("create i18n file fail", err)
	}
	defer file.Close()
	_, err = file.WriteString(b.String())
	if err != nil {
		log.Fatal("write i18n file fail", err)
	}

	if err = exec.Command("gofmt", "-w", output).Run(); err != nil {
		log.Fatal("execute go fmt fail", err)
	}
	return
}
