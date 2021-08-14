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
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type param struct{ K, V, P string }

type Parse struct {
	Path    string
	Package string
	// Imports      []string
	Imports map[string]*param
	Funcs   []*struct {
		Name   string
		Method []*param
		Params []*param
		Result []*param
	}
}

// ParseArgs parse input args
func ParseArgs(args []string, res *[]string, index int) (err error) {
	if len(args) <= index {
		return
	}
	if strings.HasPrefix(args[index], "-") {
		index += 2
		ParseArgs(args, res, index)
		return
	}
	var f os.FileInfo
	if f, err = os.Stat(args[index]); err != nil {
		return
	}
	if f.IsDir() {
		if !strings.HasSuffix(args[index], "/") {
			args[index] += "/"
		}
		var fs []os.FileInfo
		if fs, err = ioutil.ReadDir(args[index]); err != nil {
			return
		}
		for _, f = range fs {
			path, _ := filepath.Abs(args[index] + f.Name())
			args = append(args, path)
		}
	} else {
		if strings.HasSuffix(args[index], ".go") &&
			!strings.HasSuffix(args[index], "_test.go") {
			*res = append(*res, args[index])
		}
	}
	index++
	return ParseArgs(args, res, index)
}

// ParseFile parse files to Parse structs
func ParseFile(files ...string) (parses []*Parse, err error) {
	for _, file := range files {
		var (
			astFile *ast.File
			fSet    = token.NewFileSet()
			parse   = &Parse{
				Imports: make(map[string]*param),
			}
		)
		if astFile, err = parser.ParseFile(fSet, file, nil, 0); err != nil {
			return
		}
		if astFile.Name != nil {
			parse.Path = file
			parse.Package = astFile.Name.Name
		}
		for _, decl := range astFile.Decls {
			switch decl.(type) {
			case *ast.GenDecl:
				if specs := decl.(*ast.GenDecl).Specs; len(specs) > 0 {
					parse.Imports = parseImports(specs)
				}
			case *ast.FuncDecl:
				var (
					dec       = decl.(*ast.FuncDecl)
					parseFunc = &struct {
						Name                   string
						Method, Params, Result []*param
					}{Name: dec.Name.Name}
				)
				if dec.Recv != nil {
					parseFunc.Method = parserParams(dec.Recv.List)
				}
				if dec.Type.Params != nil {
					parseFunc.Params = parserParams(dec.Type.Params.List)
				}
				if dec.Type.Results != nil {
					parseFunc.Result = parserParams(dec.Type.Results.List)
				}
				parse.Funcs = append(parse.Funcs, parseFunc)
			}
		}
		parses = append(parses, parse)
	}
	return
}

// parserParams parse ast field to params
func parserParams(fields []*ast.Field) (params []*param) {
	for _, field := range fields {
		p := &param{}
		p.V = parseType(field.Type)
		if field.Names == nil {
			params = append(params, p)
		}
		for _, name := range field.Names {
			sp := &param{}
			sp.K = name.Name
			sp.V = p.V
			sp.P = p.P
			params = append(params, sp)
		}
	}
	return
}

// parseType parse ast expr to string
func parseType(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.Ident:
		return expr.(*ast.Ident).Name
	case *ast.StarExpr:
		return "*" + parseType(expr.(*ast.StarExpr).X)
	case *ast.ArrayType:
		return "[" + parseType(expr.(*ast.ArrayType).Len) + "]" + parseType(expr.(*ast.ArrayType).Elt)
	case *ast.SelectorExpr:
		return parseType(expr.(*ast.SelectorExpr).X) + "." + expr.(*ast.SelectorExpr).Sel.Name
	case *ast.MapType:
		return "map[" + parseType(expr.(*ast.MapType).Key) + "]" + parseType(expr.(*ast.MapType).Value)
	case *ast.StructType:
		return "struct{}"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		var (
			pTemp string
			rTemp string
		)
		pTemp = parseFuncType(pTemp, expr.(*ast.FuncType).Params)
		if expr.(*ast.FuncType).Results != nil {
			rTemp = parseFuncType(rTemp, expr.(*ast.FuncType).Results)
			return fmt.Sprintf("func(%s) (%s)", pTemp, rTemp)
		}
		return fmt.Sprintf("func(%s)", pTemp)
	case *ast.ChanType:
		return fmt.Sprintf("make(chan %s)", parseType(expr.(*ast.ChanType).Value))
	case *ast.Ellipsis:
		return parseType(expr.(*ast.Ellipsis).Elt)
	}
	return ""
}

// parseFuncType parse func type
func parseFuncType(temp string, data *ast.FieldList) string {
	var params = parserParams(data.List)
	for i, param := range params {
		if i == 0 {
			temp = param.K + " " + param.V
			continue
		}
		t := param.K + " " + param.V
		temp = fmt.Sprintf("%s, %s", temp, t)
	}
	return temp
}

// parseImports parse import packages
func parseImports(specs []ast.Spec) (params map[string]*param) {
	params = make(map[string]*param)
	for _, spec := range specs {
		switch spec.(type) {
		case *ast.ImportSpec:
			p := &param{V: strings.Replace(spec.(*ast.ImportSpec).Path.Value, "\"", "", -1)}
			if spec.(*ast.ImportSpec).Name != nil {
				p.K = spec.(*ast.ImportSpec).Name.Name
				params[p.K] = p
			} else {
				vs := strings.Split(p.V, "/")
				params[vs[len(vs)-1]] = p
			}
		}
	}
	return
}
