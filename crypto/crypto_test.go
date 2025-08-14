package crypto_test

import (
	"testing"

	"github.com/designinlife/slib/crypto"
	"github.com/stretchr/testify/assert"
)

func TestMD5Bytes(t *testing.T) {
	s := crypto.MD5Bytes([]byte("hello"))
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", s)
}

func TestMD5String(t *testing.T) {
	s := crypto.MD5String("hello")
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", s)
}

func TestAES256Encrypt(t *testing.T) {
	v1, err := crypto.AES256Encrypt("hello", []byte("1234567890123456"))
	assert.NoError(t, err)
	t.Log(v1)
}

func TestAES256Decrypt(t *testing.T) {
	v1, err := crypto.AES256Decrypt("UNp5+Qaxv66WLGqxr89ix7NYMCONGjU2rFv3scbASsk7", []byte("1234567890123456"))
	assert.NoError(t, err)
	assert.Equal(t, "hello", v1)

	v2, err := crypto.AES256Decrypt("I1rYJMc5QWLehI6y2L4Q6onc48zhPg8MDtIyPVz6JCgh", []byte("1234567890123456"))
	assert.NoError(t, err)
	assert.Equal(t, "hello", v2)
}

func TestChaCha20Encrypt(t *testing.T) {
	bCipherText, nonce, err := crypto.ChaCha20Encrypt([]byte("k5/qzOMicmi9Osh,ttl+4G*zXUWCP?3O"), []byte("hello"), nil)
	assert.NoError(t, err)
	t.Log(string(bCipherText))
	t.Log(string(nonce))

	bPlainText, err := crypto.ChaCha20Decrypt([]byte("k5/qzOMicmi9Osh,ttl+4G*zXUWCP?3O"), bCipherText, nonce, nil)
	assert.NoError(t, err)
	t.Log(string(bPlainText))
}
