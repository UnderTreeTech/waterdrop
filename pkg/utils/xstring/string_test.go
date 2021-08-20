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

package xstring

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubstr(t *testing.T) {
	assert.Equal(t, "waterdrop", Substr("Hello waterdrop!", 6, 15))
	assert.Equal(t, "waterdrop!", Substr("Hello waterdrop!", 6, -1))
	assert.Equal(t, "Hello waterdrop!", Substr("Hello waterdrop!", 0, -1))
	assert.Equal(t, "Hello waterdrop!", Substr("Hello waterdrop!", 0, len("Hello waterdrop!")+1))
	assert.NotEqual(t, "Hello waterdrop!", Substr("Hello waterdrop!", -1, len("Hello waterdrop!")+1))
	assert.Equal(t, "", Substr("Hello waterdrop!", -1, len("Hello waterdrop!")+1))
	assert.Equal(t, "", Substr("Hello waterdrop!", len("Hello waterdrop!")+1, len("Hello waterdrop!")+1))
}

func TestReverse(t *testing.T) {
	assert.Equal(t, "!pordretaw olleH", Reverse("Hello waterdrop!"))
	assert.Equal(t, "World Peace", Reverse("ecaeP dlroW"))
}

func TestGenerateUUID(t *testing.T) {
	assert.Equal(t, 36, len(GenerateUUID()))
}

func TestStringToBytes(t *testing.T) {
	for i := 0; i < 100; i++ {
		s := RandomString(64)
		if !bytes.Equal([]byte(s), StringToBytes(s)) {
			assert.Fail(t, "don't match")
		}
	}
}

func TestBytesToString(t *testing.T) {
	data := make([]byte, 1024)
	for i := 0; i < 100; i++ {
		rand.Read(data)
		if string(data) != BytesToString(data) {
			assert.Fail(t, "don't match")
		}
	}
}

func TestStripContentType(t *testing.T) {
	assert.Equal(t, StripContentType("application/json;charset=utf-8"), "application/json")
	assert.Equal(t, StripContentType("application/json"), "application/json")
}

func TestGetLocaleLng(t *testing.T) {
	assert.Equal(t, _defalutLng, GetLocaleLng("*"))
	assert.Equal(t, _defalutLng, GetLocaleLng(""))
	assert.Equal(t, "fr", GetLocaleLng("fr-CH, fr;q=0.9"))
	assert.Equal(t, "en", GetLocaleLng("en;q=0.8"))
	assert.Equal(t, _defalutLng, GetLocaleLng("*;q=0.5"))
	assert.Equal(t, "zh", GetLocaleLng(" zh-CN,zh;q=0.5"))
}

// BenchmarkRandomString bench generate random length string
func BenchmarkRandomString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = RandomString(16)
		}
	})
}
