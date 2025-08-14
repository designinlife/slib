package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

func ChaCha20Encrypt(plaintext string, key []byte) (string, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err1 := io.ReadFull(rand.Reader, nonce); err1 != nil {
		return "", err1
	}

	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func ChaCha20Decrypt(ciphertext string, key []byte) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	nonce := ciphertextBytes[:aead.NonceSize()]
	ciphertextBytes = ciphertextBytes[aead.NonceSize():]

	plaintext, err := aead.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
