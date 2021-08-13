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

package xslice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	s1 := []string{"hello", "world", "hello+", "waterdrop"}
	s2 := []string{"hello", "world", "github", "waterdrop"}
	s3 := []string{"xstring"}

	assert.Equal(t, true, ContainString(s1, "hello"))
	assert.Equal(t, false, ContainString(s2, "hello+"))
	assert.ElementsMatch(t, []string{"hello+"}, DiffString(s1, s2))
	assert.ElementsMatch(t, []string{"github"}, DiffString(s2, s1))
	assert.ElementsMatch(t, []string{"hello", "world", "waterdrop"}, IntersectString(s1, s2))
	assert.ElementsMatch(t, []string{"hello", "world", "hello+", "waterdrop", "hello", "world", "github", "waterdrop"}, MergeString(s1, s2))
	assert.ElementsMatch(t, []string{"xstring"}, MergeString(s3))
	assert.ElementsMatch(t, []string{"hello", "world", "hello+", "waterdrop", "hello", "world", "github", "waterdrop", "xstring"}, MergeString(s1, s2, s3))
	assert.ElementsMatch(t, []string{"hello", "world", "hello+", "github", "waterdrop"}, UniqueString(MergeString(s1, s2)))
	assert.Equal(t, []string{"hello", "hello+", "waterdrop", "world"}, SortString(s1))
	assert.ElementsMatch(t, []string{"hello", "world", "hello+"}, RemoveString(s1, "waterdrop"))
	assert.ElementsMatch(t, []string{"hello", "world", "hello+", "waterdrop"}, RemoveString(s1, "waterdrop+"))
	assert.ElementsMatch(t, []string{"waterdrop", "hello+", "world", "hello"}, ReverseString(s1))
}
