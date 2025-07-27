package crypto

func SHA1Bytes(b []byte) string {
	return encodeBytes(b, AlgorithmSHA1)
}

func SHA1String(s string) string {
	return encodeString(s, AlgorithmSHA1)
}

func SHA1File(path string) (string, error) {
	return encodeFile(path, AlgorithmSHA1)
}
