// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xcrypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"test", "098f6bcd4621d373cade4e832627b4f6"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, MD5(tt.input))
		})
	}
}

func TestMD5Short(t *testing.T) {
	assert.Equal(t, "5d41402a", MD5Short("hello"))
}

func TestSHA1(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, SHA1(tt.input))
		})
	}
}

func TestSHA256(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, SHA256(tt.input))
		})
	}
}

func TestSHA512(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{"", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, SHA512(tt.input))
		})
	}
}

func TestHmacMD5(t *testing.T) {
	result := HmacMD5("hello", "key")
	assert.NotEmpty(t, result)
	assert.Equal(t, 32, len(result)) // MD5 hex is 32 chars
}

func TestHmacSHA1(t *testing.T) {
	result := HmacSHA1("hello", "key")
	assert.NotEmpty(t, result)
	assert.Equal(t, 40, len(result)) // SHA1 hex is 40 chars
}

func TestHmacSHA256(t *testing.T) {
	result := HmacSHA256("hello", "key")
	assert.NotEmpty(t, result)
	assert.Equal(t, 64, len(result)) // SHA256 hex is 64 chars
}

func TestBase64EncodeDecode(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"hello world"},
		{""},
		{"special chars: !@#$%^&*()"},
		{"中文测试"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			encoded := Base64Encode(tt.input)
			decoded, err := Base64Decode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decoded)
		})
	}
}

func TestBase64URLEncodeDecode(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"hello world"},
		{""},
		{"special chars: !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			encoded := Base64URLEncode(tt.input)
			decoded, err := Base64URLDecode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decoded)
		})
	}
}

func TestBase64RawURLEncodeDecode(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"hello world"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			encoded := Base64RawURLEncode(tt.input)
			decoded, err := Base64RawURLDecode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decoded)
		})
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	key := "1234567890123456" // 16 bytes for AES-128
	plaintext := "hello world"

	encrypted, err := AESEncrypt(key, plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := AESDecrypt(key, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESGCMEncryptDecrypt(t *testing.T) {
	key := "1234567890123456" // 16 bytes for AES-128
	plaintext := "hello world"

	encrypted, err := AESGCMEncrypt(key, plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := AESGCMDecrypt(key, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestRSAEncryptDecrypt(t *testing.T) {
	priv, pub, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)

	plaintext := "hello world"

	encrypted, err := RSAEncrypt(pub, plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := RSADecrypt(priv, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestRSAKeyPairPEM(t *testing.T) {
	priv, pub, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)

	// Public key
	pubPEM, err := RSAPublicKeyToPEM(pub)
	assert.NoError(t, err)
	assert.Contains(t, pubPEM, "RSA PUBLIC KEY")

	parsedPub, err := RSAPEMToPublicKey(pubPEM)
	assert.NoError(t, err)
	assert.Equal(t, pub, parsedPub)

	// Private key
	privPEM, err := RSAPrivateKeyToPEM(priv)
	assert.NoError(t, err)
	assert.Contains(t, privPEM, "RSA PRIVATE KEY")

	parsedPriv, err := RSAPEMToPrivateKey(privPEM)
	assert.NoError(t, err)
	assert.Equal(t, priv.D, parsedPriv.D)
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short", 10},
		{"medium", 50},
		{"long", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)
			assert.Equal(t, tt.length, len(result))
			assert.Regexp(t, "^[a-zA-Z0-9]+$", result)
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	bytes, err := GenerateRandomBytes(32)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(bytes))
}

func TestXOR(t *testing.T) {
	original := "hello"
	key := "key"

	encrypted := XOR(original, key)
	assert.NotEqual(t, original, encrypted)

	decrypted := XOR(encrypted, key)
	assert.Equal(t, original, decrypted)
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"abc", "cba"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, Reverse(tt.input))
		})
	}
}

func TestCaesar(t *testing.T) {
	tests := []struct {
		input    string
		shift    int
		expected string
	}{
		{"hello", 1, "ifmmp"},
		{"hello", 0, "hello"},
		{"hello", 26, "hello"},
		{"HELLO", 1, "IFMMP"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Caesar(tt.input, tt.shift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestROT13(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "uryyb"},
		{"uryyb", "hello"},
		{"HELLO", "URYYB"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ROT13(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase64DecodeError(t *testing.T) {
	_, err := Base64Decode("invalid base64!!!")
	assert.Error(t, err)
}

func TestBase64URLDecodeError(t *testing.T) {
	_, err := Base64URLDecode("invalid base64!!!")
	assert.Error(t, err)
}

func TestBase64RawURLDecodeError(t *testing.T) {
	_, err := Base64RawURLDecode("invalid base64!!!")
	assert.Error(t, err)
}

func TestAESDecryptError(t *testing.T) {
	_, err := AESDecrypt("1234567890123456", "invalid ciphertext")
	assert.Error(t, err)
}

func TestAESGCMDecryptError(t *testing.T) {
	_, err := AESGCMDecrypt("1234567890123456", "invalid ciphertext")
	assert.Error(t, err)
}
