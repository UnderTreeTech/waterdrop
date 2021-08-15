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

func TestRsa(t *testing.T) {
	content := "High performance go framework waterdrop"
	config := &RsaConfig{
		PrivateKeyPath: "./pem/rsa_private_key.pem",
		PublicKeyPath:  "./pem/rsa_public_key.pem",
	}
	rsa, err := NewRsaCrypt(config)
	assert.Equal(t, err, nil)

	encrypt, err := rsa.Encrypt([]byte(content))
	assert.Equal(t, err, nil)
	decrypt, err := rsa.Decrypt(encrypt)
	assert.Equal(t, err, nil)
	assert.Equal(t, content, string(decrypt))

	hexEncrypt, err := rsa.EncryptToString([]byte(content), HEX)
	assert.Equal(t, err, nil)
	hexDecrypt, err := rsa.DecryptFromString(hexEncrypt, HEX)
	assert.Equal(t, err, nil)
	assert.Equal(t, content, string(hexDecrypt))

	base64Encrypt, err := rsa.EncryptToString([]byte(content), Base64)
	assert.Equal(t, err, nil)
	base64Decrypt, err := rsa.DecryptFromString(base64Encrypt, Base64)
	assert.Equal(t, err, nil)
	assert.Equal(t, content, string(base64Decrypt))

	byteEncrypt, err := rsa.EncryptToString([]byte(content), PlainText)
	assert.Equal(t, err, nil)
	byteDecrypt, err := rsa.DecryptFromString(byteEncrypt, PlainText)
	assert.Equal(t, err, nil)
	assert.Equal(t, content, string(byteDecrypt))
}
