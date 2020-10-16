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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/testgen/common"
)

func genTest(parses []*common.Parse) (err error) {
	for _, p := range parses {
		if err = genUTTest(p); err != nil {
			break
		}
	}
	return
}

func genUTTest(p *common.Parse) (err error) {
	var (
		buffer bytes.Buffer
		impts  = strings.Join([]string{
			`"context"`,
			`"testing"`,
			`. "github.com/smartystreets/goconvey/convey"`,
		}, "\n\t")
		content []byte
	)
	filename := strings.Replace(p.Path, ".go", "_test.go", -1)
	if _, err = os.Stat(filename); (genFunc == "" && err == nil) ||
		(err != nil && os.IsExist(err)) {
		err = nil
		return
	}
	for _, impt := range p.Imports {
		impts += "\n\t\"" + impt.V + "\""
	}
	if genFunc == "" {
		buffer.WriteString(fmt.Sprintf(tpPackage, p.Package))
		buffer.WriteString(fmt.Sprintf(tpImport, impts))
	}
	for _, parseFunc := range p.Funcs {
		if genFunc != "" && genFunc != parseFunc.Name {
			continue
		}
		var (
			methodK string
			tpVars  string
			vars    []string
			val     []string
			notice  = "Then "
			reset   string
		)
		if method := common.ConvertMethod(p.Path); method != "" {
			methodK = method + "."
		}
		tpTestFuncs := fmt.Sprintf(tpTestFunc, strings.Title(p.Package), parseFunc.Name, "", parseFunc.Name, "%s", "%s", "%s")
		tpTestFuncBeCall := methodK + parseFunc.Name + "(%s)\n\t\t\tConvey(\"%s\", func() {"
		if parseFunc.Result == nil {
			tpTestFuncBeCall = fmt.Sprintf(tpTestFuncBeCall, "%s", "No return values")
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", tpTestFuncBeCall, "%s")
		}
		for k, res := range parseFunc.Result {
			if res.K == "" {
				res.K = fmt.Sprintf("p%d", k+1)
			}
			var so string
			if res.V == "error" {
				res.K = "err"
				so = fmt.Sprintf("\tSo(%s, ShouldBeNil)", res.K)
				notice += "err should be nil."
			} else {
				so = fmt.Sprintf("\tSo(%s, ShouldNotBeNil)", res.K)
				val = append(val, res.K)
			}
			if len(parseFunc.Result) <= k+1 {
				if len(val) != 0 {
					notice += strings.Join(val, ",") + " should not be nil."
				}
				tpTestFuncBeCall = fmt.Sprintf(tpTestFuncBeCall, "%s", notice)
				res.K += " := " + tpTestFuncBeCall
			} else {
				res.K += ", %s"
			}
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", res.K+"\n\t\t\t%s", "%s")
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", "%s", so, "%s")
		}
		if parseFunc.Params == nil {
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", "", "%s")
		}
		for k, pType := range parseFunc.Params {
			if pType.K == "" {
				pType.K = fmt.Sprintf("a%d", k+1)
			}
			var (
				init   string
				params = pType.K
			)
			switch {
			case strings.HasPrefix(pType.V, "context"):
				init = params + " = context.Background()"
			case strings.HasPrefix(pType.V, "[]byte"):
				init = params + " = " + pType.V + "(\"\")"
			case strings.HasPrefix(pType.V, "[]"):
				init = params + " = " + pType.V + "{}"
			case strings.HasPrefix(pType.V, "int") ||
				strings.HasPrefix(pType.V, "uint") ||
				strings.HasPrefix(pType.V, "float") ||
				strings.HasPrefix(pType.V, "double"):
				init = params + " = " + pType.V + "(0)"
			case strings.HasPrefix(pType.V, "string"):
				init = params + " = \"\""
			case strings.Contains(pType.V, "*xsql.Tx"):
				init = params + ",_ = " + methodK + "BeginTran(c)"
				reset += "\n\t" + params + ".Commit()"
			case strings.HasPrefix(pType.V, "*"):
				init = params + " = " + strings.Replace(pType.V, "*", "&", -1) + "{}"
			case strings.Contains(pType.V, "chan"):
				init = params + " = " + pType.V
			case pType.V == "time.Time":
				init = params + " = time.Now()"
			case strings.Contains(pType.V, "chan"):
				init = params + " = " + pType.V
			default:
				init = params + " " + pType.V
			}
			vars = append(vars, "\t\t"+init)
			if len(parseFunc.Params) > k+1 {
				params += ", %s"
			}
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "%s", params, "%s")
		}
		if len(vars) > 0 {
			tpVars = fmt.Sprintf(tpVar, strings.Join(vars, "\n\t"))
		}
		tpTestFuncs = fmt.Sprintf(tpTestFuncs, tpVars, "%s")
		if reset != "" {
			tpTestResets := fmt.Sprintf(tpTestReset, reset)
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, tpTestResets)
		} else {
			tpTestFuncs = fmt.Sprintf(tpTestFuncs, "")
		}
		buffer.WriteString(tpTestFuncs)
	}
	var (
		file *os.File
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	)
	if file, err = os.OpenFile(filename, flag, 0644); err != nil {
		return
	}
	if genFunc == "" {
		content, _ = common.GoImport(filename, buffer.Bytes())
	} else {
		content = buffer.Bytes()
	}
	if _, err = file.Write(content); err != nil {
		return
	}
	if err = file.Close(); err != nil {
		return
	}
	return
}
