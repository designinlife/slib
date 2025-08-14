package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// ChaCha20Encrypt 使用 ChaCha20-Poly1305 加密数据。
// 它返回密文和用于解密的 nonce。
//
// 密钥必须是 32 字节长。
// associatedData 是可选的，用于认证而不加密的数据。
func ChaCha20Encrypt(key, plaintext, associatedData []byte) ([]byte, []byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", chacha20poly1305.KeySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ChaCha20-Poly1305 cipher: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// aead.Seal 的签名是 Seal(dst, nonce, plaintext, additionalData []byte) []byte
	// 它将加密的 plaintext 附加到 dst (如果 dst 为 nil，则创建一个新的切片)
	// 并返回结果。
	ciphertext := aead.Seal(nil, nonce, plaintext, associatedData)

	return ciphertext, nonce, nil
}

// ChaCha20Decrypt 使用 ChaCha20-Poly1305 解密数据。
//
// 密钥必须是 32 字节长。
// nonce 必须是加密时使用的那个 nonce。
// associatedData 必须是加密时使用的那个 associatedData，或者 nil。
func ChaCha20Decrypt(key, ciphertext, nonce, associatedData []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", chacha20poly1305.KeySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305 cipher: %w", err)
	}

	if len(nonce) != aead.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", aead.NonceSize(), len(nonce))
	}

	// aead.Open 的签名是 Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error)
	// 它解密 ciphertext，验证 authenticity，如果成功则将明文附加到 dst 并返回。
	// 如果认证失败（数据被篡改），则返回错误。
	plaintext, err := aead.Open(nil, nonce, ciphertext, associatedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt or authenticate data: %w", err)
	}

	return plaintext, nil
}
