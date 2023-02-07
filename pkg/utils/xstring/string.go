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

package xstring

import (
	"math/rand"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

const (
	_defalutLng  = "en"
	_lngWildCard = "*"
	_letters     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

//get locale language
func GetLocaleLng(lng string) string {
	// Multiple types, weighted with the quality value syntax:
	//Accept-Language: fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5, en-US,en;q=0.5, en, *
	locale := strings.Split(strings.Split(lng, ";")[0], "-")[0]
	locale = strings.TrimSpace(locale)

	//lng default to en if it doesn't have Accept-Language header or accept any language
	if "" == locale || _lngWildCard == locale {
		locale = _defalutLng
	}

	return strings.ToLower(locale)
}

// strip content-type
// application/json;charset=utf-8
func StripContentType(contentType string) string {
	i := strings.Index(contentType, ";")
	if i != -1 {
		contentType = contentType[:i]
	}
	return contentType
}

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) (b []byte) {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// generate random string by len
func RandomString(length int) string {
	sb := strings.Builder{}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		sb.WriteByte(_letters[rand.Intn(len(_letters))])
	}

	return sb.String()
}

// GenerateUUID creates a new random UUID4 and returns it as a string or panics.
func GenerateUUID() string {
	return uuid.NewString()
}

// Substr returns runes between start and stop [start, stop) regardless of the chars are ascii or utf8.
// if start < 0 or start < length, it returns empty string
// if stop < 0 or stop > length, it returns rs[start:], otherwise it returns rs[start:stop]
func Substr(str string, start, stop int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return ""
	}

	if stop < 0 || stop > length {
		return string(rs[start:])
	}

	return string(rs[start:stop])
}

// Reverse reverses s.
func Reverse(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}
