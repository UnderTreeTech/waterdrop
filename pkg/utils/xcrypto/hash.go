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
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"io"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
)

// Hash hash input data using assigned hash mode
func Hash(src string, mode HashMode) (dst []byte) {
	dst = hashEncrypt(xstring.StringToBytes(src), mode)
	return
}

// HashToString hash input to assigned encode mode.
func HashToString(src string, mode HashMode, encode Encode) (dst string, err error) {
	encrypt := Hash(src, mode)
	dst, err = EncodeToString(encrypt, encode)
	return
}

// hashEncrypt encrypt input using assigned hash mode
// hash mode: md5, sha1, sha256, sha512, default md5
func hashEncrypt(data []byte, mode HashMode) (hashed []byte) {
	f := getHashFunc(mode)
	hh := f()
	hh.Write(data)
	hashed = hh.Sum(nil)
	return
}

// getHashFunc gets the crypto hash func
func getHashFunc(mode HashMode) (f func() hash.Hash) {
	switch mode {
	case SHA1:
		return sha1.New
	case SHA256:
		return sha256.New
	case SHA512:
		return sha512.New
	case MD5:
		return md5.New
	default:
		return md5.New
	}
}

// HmacSHA256 returns HMAC bytes for src with the given key
func HmacSHA256(key []byte, src string) (dst []byte) {
	h := hmac.New(getHashFunc(SHA256), key)
	io.WriteString(h, src)
	dst = h.Sum(nil)
	return
}

// HmacSHA256ToString returns hmac hash string with the given key and assigned encode mode
// hash mode: md5, sha1, sha256, sha512, default md5
func HmacSHA256ToString(key []byte, src string, encode Encode) (dst string, err error) {
	encrypt := HmacSHA256(key, src)
	dst, err = EncodeToString(encrypt, encode)
	return
}
