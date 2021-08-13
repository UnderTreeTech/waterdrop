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

package xcrypto

import (
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"
)

//Hash for crypto Hash
type HashMode uint

const (
	MD5 HashMode = iota
	SHA1
	SHA256
	SHA512
)

//Encode defines the type of bytes encoded to string
type Encode uint

const (
	PlainText Encode = iota // plain text, no encode just string
	HEX                     // hex encode
	Base64                  // base64 encode
)

// ErrInvalidEncodeMode unsupported encode mode
var ErrInvalidEncodeMode = errors.New("unsupported encode mode")

// DecodeString decodes string data to bytes in designed encoded type
func DecodeString(data string, encodedType Encode) (dst []byte, err error) {
	switch encodedType {
	case PlainText:
		dst = xstring.StringToBytes(data)
	case HEX:
		dst, err = hex.DecodeString(data)
	case Base64:
		dst, err = base64.StdEncoding.DecodeString(data)
	default:
		return dst, ErrInvalidEncodeMode
	}
	return
}

//EncodeToString encodes data to string with encode type
func EncodeToString(data []byte, encodeType Encode) (dst string, err error) {
	switch encodeType {
	case HEX:
		dst = hex.EncodeToString(data)
	case Base64:
		dst = base64.StdEncoding.EncodeToString(data)
	case PlainText:
		dst = xstring.BytesToString(data)
	default:
		return "", ErrInvalidEncodeMode
	}
	return
}
