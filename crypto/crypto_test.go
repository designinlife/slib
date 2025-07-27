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
