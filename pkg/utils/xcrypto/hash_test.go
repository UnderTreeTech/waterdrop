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

package xcrypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	content := "123456"
	md5Str := "e10adc3949ba59abbe56e057f20f883e"
	sha1Str := "7c4a8d09ca3762af61e59520943dc26494f8941b"
	sha256Str := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"
	sha512Str := "ba3253876aed6bc22d4a6ff53d8406c6ad864195ed144ab5c87621b6c233b548baeae6956df346ec8c17f5ea10f35ee3cbc514797ed7ddd3145464e2a0bab413"

	encryptMd5Str, _ := HashToString(content, MD5, HEX)
	assert.Equal(t, md5Str, encryptMd5Str)
	encryptSha1Str, _ := HashToString(content, SHA1, HEX)
	assert.Equal(t, sha1Str, encryptSha1Str)
	encryptSha256Str, _ := HashToString(content, SHA256, HEX)
	assert.Equal(t, sha256Str, encryptSha256Str)
	encryptSha512Str, _ := HashToString(content, SHA512, HEX)
	assert.Equal(t, sha512Str, encryptSha512Str)
}

func TestHmacSHA256(t *testing.T) {
	content := "123456"
	key := "waterdrop"
	encrypt := "74a05d4cf0a06caafa4e3a8cbe3bb04a1476178aed8cb53dc109edd8ae3c3f30"

	dst, err := HmacSHA256ToString([]byte(key), content, HEX)
	assert.Nil(t, err)
	assert.Equal(t, encrypt, dst)
}
