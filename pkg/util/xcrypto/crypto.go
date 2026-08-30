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
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"hash"
)

// MD5 calculates the MD5 hash of the input string.
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// MD5Short returns the first 8 characters of MD5 hash.
func MD5Short(s string) string {
	return MD5(s)[:8]
}

// SHA1 calculates the SHA1 hash of the input string.
func SHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 calculates the SHA256 hash of the input string.
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA512 calculates the SHA512 hash of the input string.
func SHA512(s string) string {
	h := sha512.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// HmacMD5 calculates the HMAC-MD5 hash.
func HmacMD5(message, key string) string {
	return HmacWithHash(md5.New, message, key)
}

// HmacSHA1 calculates the HMAC-SHA1 hash.
func HmacSHA1(message, key string) string {
	return HmacWithHash(sha1.New, message, key)
}

// HmacSHA256 calculates the HMAC-SHA256 hash.
func HmacSHA256(message, key string) string {
	return HmacWithHash(sha256.New, message, key)
}

// HmacWithHash calculates HMAC with the specified hash function.
func HmacWithHash(hashFunc func() hash.Hash, message, key string) string {
	mac := hmac.New(hashFunc, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// Base64Encode encodes the input to base64.
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode decodes the base64 input.
func Base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Base64URLEncode encodes the input to base64 URL.
func Base64URLEncode(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// Base64URLDecode decodes the base64 URL input.
func Base64URLDecode(s string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Base64RawURLEncode encodes the input to base64 raw URL.
func Base64RawURLEncode(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// Base64RawURLDecode decodes the base64 raw URL input.
func Base64RawURLDecode(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AESEncrypt encrypts the input using AES-CBC with PKCS7 padding.
func AESEncrypt(key, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	plaintextBytes := []byte(plaintext)
	plaintextBytes = pkcs7Pad(plaintextBytes, aes.BlockSize)

	ciphertext := make([]byte, aes.BlockSize+len(plaintextBytes))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(ciphertext[aes.BlockSize:], plaintextBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt decrypts the input using AES-CBC with PKCS7 padding.
func AESDecrypt(key, ciphertext string) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	stream := cipher.NewCBCDecrypter(block, iv)
	stream.CryptBlocks(ciphertextBytes, ciphertextBytes)

	ciphertextBytes, err = pkcs7Unpad(ciphertextBytes, aes.BlockSize)
	if err != nil {
		return "", err
	}

	return string(ciphertextBytes), nil
}

// AESGCMEncrypt encrypts the input using AES-GCM.
func AESGCMEncrypt(key, plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESGCMDecrypt decrypts the input using AES-GCM.
func AESGCMDecrypt(key, ciphertext string) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// RSAEncrypt encrypts the plaintext using RSA public key.
func RSAEncrypt(pub *rsa.PublicKey, plaintext string) (string, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecrypt decrypts the ciphertext using RSA private key.
func RSADecrypt(priv *rsa.PrivateKey, ciphertext string) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertextBytes)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// GenerateRSAKeyPair generates an RSA key pair.
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// RSAPublicKeyToPEM converts RSA public key to PEM format.
func RSAPublicKeyToPEM(pub *rsa.PublicKey) (string, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}

	pubBytes := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	}

	return string(pem.EncodeToMemory(pubBytes)), nil
}

// RSAPrivateKeyToPEM converts RSA private key to PEM format.
func RSAPrivateKeyToPEM(priv *rsa.PrivateKey) (string, error) {
	privBytes := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}

	return string(pem.EncodeToMemory(privBytes)), nil
}

// RSAPEMToPublicKey converts PEM to RSA public key.
func RSAPEMToPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}

// RSAPEMToPrivateKey converts PEM to RSA private key.
func RSAPEMToPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		var buf [1]byte
		_, _ = rand.Read(buf[:])
		b[i] = charset[int(buf[0])%len(charset)]
	}
	return string(b)
}

// GenerateRandomBytes generates random bytes of the specified length.
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// XOR encrypts/decrypts the input using XOR with the key.
func XOR(data, key string) string {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return string(result)
}

// Reverse reverses the input string.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Caesar encrypts the input using Caesar cipher with the given shift.
func Caesar(s string, shift int) string {
	result := make([]rune, len([]rune(s)))
	for i, r := range s {
		if r >= 'a' && r <= 'z' {
			result[i] = 'a' + (r-'a'+rune(shift))%26
		} else if r >= 'A' && r <= 'Z' {
			result[i] = 'A' + (r-'A'+rune(shift))%26
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// ROT13 encrypts/decrypts the input using ROT13.
func ROT13(s string) string {
	return Caesar(s, 13)
}

// pkcs7Pad adds PKCS7 padding to the input.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS7 padding from the input.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}
