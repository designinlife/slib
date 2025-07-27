package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/designinlife/slib/fs"
)

type Algorithm int

const (
	AlgorithmMD5 Algorithm = iota + 1
	AlgorithmSHA1
	AlgorithmSHA224
	AlgorithmSHA256
	AlgorithmSHA384
	AlgorithmSHA512
)

func getHashEncoder(algorithm Algorithm) hash.Hash {
	switch algorithm {
	case AlgorithmMD5:
		return md5.New()
	case AlgorithmSHA1:
		return sha1.New()
	case AlgorithmSHA224:
		return sha256.New224()
	case AlgorithmSHA256:
		return sha256.New()
	case AlgorithmSHA384:
		return sha512.New384()
	case AlgorithmSHA512:
		return sha512.New()
	default:
		return md5.New()
	}
}

func encodeBytes(b []byte, algorithm Algorithm) string {
	h := getHashEncoder(algorithm)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func encodeString(s string, algorithm Algorithm) string {
	return encodeBytes([]byte(s), algorithm)
}

func encodeFile(path string, algorithm Algorithm) (string, error) {
	if !fs.IsFile(path) {
		return "", fmt.Errorf("file does not exist. (%s)", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := getHashEncoder(algorithm)
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
