package utils

import (
	"math/rand"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

const (
	_defalutLng  = "en"
	_lngWildCard = "*"
	_letters     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-_"
)

//get locale language
func GetLocaleLng(lng string) string {
	// Multiple types, weighted with the quality value syntax:
	//Accept-Language: fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5, en-US,en;q=0.5, en, *
	locale := strings.Split(strings.Split(lng, ";")[0], "-")[0]

	//lng default to en if it doesn't have Accept-Language header or accept any language
	if "" == locale || _lngWildCard == locale {
		locale = _defalutLng
	}

	return strings.ToLower(locale)
}

// get current unix time
func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
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
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
