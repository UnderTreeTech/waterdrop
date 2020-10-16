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

var (
	tpPackage   = "package %s\n\n"
	tpImport    = "import (\n\t%s\n)\n\n"
	tpVar       = "var (\n\t%s\n)\n"
	tpTestReset = "\n\t\tReset(func() {%s\n\t\t})"
	tpTestFunc  = "\nfunc Test%s%s(t *testing.T){%s\n\tConvey(\"%s\", t, func(){\n\t\t%s\tConvey(\"When everything goes positive\", func(){\n\t\t\t%s\n\t\t\t})\n\t\t})%s\n\t})\n}\n\n"
)
