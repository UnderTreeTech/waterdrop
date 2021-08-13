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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

var (
	// ErrPrivateKey indicates the invalid private key.
	ErrPrivateKey = errors.New("failed to parse private PEM block containing the key")
	// ErrPublicKey indicates the invalid public key.
	ErrPublicKey = errors.New("failed to parse public PEM block containing the key")
	// ErrNotRsaKey indicates the invalid RSA key.
	ErrNotRsaKey = errors.New("key type is not RSA")
)

type RsaCrypt struct {
	config     *RsaConfig
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

type RsaConfig struct {
	PublicKeyPath  string
	PrivateKeyPath string
}

// NewRsaCrypt init with the RSA config
func NewRsaCrypt(config *RsaConfig) (*RsaCrypt, error) {
	privateKey, err := parsePrivateKey(config.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(config.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	rsa := &RsaCrypt{
		config:     config,
		publicKey:  publicKey,
		privateKey: privateKey,
	}
	return rsa, nil
}

// Encrypt encrypts the given message with public key and return byte slice
func (rc *RsaCrypt) Encrypt(src []byte) (dst []byte, err error) {
	dst, err = rsa.EncryptPKCS1v15(rand.Reader, rc.publicKey, src)
	return
}

// EncryptToString encrypts the given message with public key and return string
func (rc *RsaCrypt) EncryptToString(src []byte, mode Encode) (dst string, err error) {
	encrypt, err := rc.Encrypt(src)
	if err != nil {
		return "", err
	}

	dst, err = EncodeToString(encrypt, mode)
	return
}

// Decrypt decrypts input slice using private key and return byte slice
// src the encrypted data with public key
func (rc *RsaCrypt) Decrypt(src []byte) (dst []byte, err error) {
	dst, err = rsa.DecryptPKCS1v15(rand.Reader, rc.privateKey, src)
	return
}

// DecryptFromString decrypts a plaintext using private key and return string
// src the encrypted data with public key
// mode the encode type of encrypted data,such as Base64,HEX or Plain String
func (rc *RsaCrypt) DecryptFromString(src string, mode Encode) (dst []byte, err error) {
	decodeData, err := DecodeString(src, mode)
	if err != nil {
		return
	}
	dst, err = rc.Decrypt(decodeData)
	return
}

// GenRsaKey return publicKey, privateKey, error
// bits: 1024, 2048, 3072, 4096
func GenRsaKey(bits int) (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}
	privateStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block1 := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateStream,
	}
	privateKeyBytes := pem.EncodeToMemory(&block1)

	publicKey := privateKey.PublicKey
	publicStream, err := x509.MarshalPKIXPublicKey(&publicKey)
	block2 := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicStream,
	}
	publicKeyBytes := pem.EncodeToMemory(&block2)

	return string(publicKeyBytes), string(privateKeyBytes), nil
}

// parsePrivateKey parses private key bytes to rsa PrivateKey
func parsePrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyContent)
	if block == nil {
		return nil, ErrPrivateKey
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// parsePublicKey parses public key bytes to rsa PublicKey
func parsePublicKey(path string) (*rsa.PublicKey, error) {
	publicKeyContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(publicKeyContent)
	if block == nil {
		return nil, ErrPublicKey
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, ErrNotRsaKey
	}
}
