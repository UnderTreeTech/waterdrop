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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

var (
	// ErrPaddingSize indicates bad padding size
	ErrPaddingSize = errors.New("padding size error")
	// ErrNotFullPadding indicates bad padding size
	ErrNotFullPadding = errors.New("need a multiple of the block size")
)

// Aes aes crypt interface definition
type Aes interface {
	Encrypt(src []byte, ivs ...[]byte) (dst []byte, err error)
	EncryptToString(src []byte, mode Encode, ivs ...[]byte) (dst string, err error)
	Decrypt(src []byte, ivs ...[]byte) (dst []byte, err error)
	DecryptFromString(src string, mode Encode, ivs ...[]byte) (dst []byte, err error)
}

// AesCbcCrypt aes cbc mode
type AesCbcCrypto struct {
	block     cipher.Block
	blockSize int
	iv        []byte
}

// NewAesCbcCrypto returns a aes cbc mode crypto
func NewAesCbcCrypto(secret string, iv ...byte) (*AesCbcCrypto, error) {
	b, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return nil, err
	}

	if len(iv) == 0 {
		iv = bytes.Repeat([]byte{0x00}, b.BlockSize())
	}

	cbc := &AesCbcCrypto{
		block:     b,
		blockSize: b.BlockSize(),
		iv:        iv,
	}
	return cbc, nil
}

// Encrypt encrypt input slice using assigned ivs
func (acc *AesCbcCrypto) Encrypt(src []byte, ivs ...[]byte) (dst []byte, err error) {
	padded := pKCS7Padding(src, acc.blockSize)
	if len(padded)%acc.blockSize != 0 {
		return nil, ErrNotFullPadding
	}
	var iv []byte
	if len(ivs) > 0 {
		iv = ivs[0]
	} else {
		iv = acc.iv
	}
	bm := cipher.NewCBCEncrypter(acc.block, iv)
	dst = make([]byte, len(padded))
	bm.CryptBlocks(dst, padded)
	return
}

// EncryptToString encrypt input slice using assigned ivs and return encode encrypt string
// encode mode: Base64,HEX or Plain String
func (acc *AesCbcCrypto) EncryptToString(src []byte, mode Encode, ivs ...[]byte) (dst string, err error) {
	data, err := acc.Encrypt(src, ivs...)
	if err != nil {
		return "", err
	}

	dst, err = EncodeToString(data, mode)
	return
}

// Decrypt decrypt input slice using assigned ivs
func (acc *AesCbcCrypto) Decrypt(src []byte, ivs ...[]byte) (dst []byte, err error) {
	var iv []byte
	if len(ivs) > 0 {
		iv = ivs[0]
	} else {
		iv = acc.iv
	}
	bm := cipher.NewCBCDecrypter(acc.block, iv)
	dst = make([]byte, len(src))
	bm.CryptBlocks(dst, src)
	dst, err = pKCS7UnPadding(dst, acc.blockSize)
	if err != nil {
		return nil, err
	}
	return
}

// DecryptFromString decrypt input string using assigned encode mode
// encode mode: Base64,HEX or Plain String
func (acc *AesCbcCrypto) DecryptFromString(src string, mode Encode, ivs ...[]byte) (dst []byte, err error) {
	decodeData, err := DecodeString(src, mode)
	if err != nil {
		return
	}
	dst, err = acc.Decrypt(decodeData, ivs...)
	return
}

// pKCS7UnPadding un-padding src data to original data , adapt to PKCS5 &PKCS7
func pKCS7UnPadding(src []byte, blockSize int) ([]byte, error) {
	n := len(src)
	if n == 0 {
		return src, nil
	}
	paddingNum := int(src[n-1])
	if paddingNum >= n || paddingNum > blockSize {
		return nil, ErrPaddingSize
	}
	return src[:(n - paddingNum)], nil
}

// pKCS7Padding adds padding data using pkcs7 rules , adapt to PKCS5 &PKCS7
func pKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}
