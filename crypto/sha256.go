package crypto

func SHA256Bytes(b []byte) string {
	return encodeBytes(b, AlgorithmSHA256)
}

func SHA256String(s string) string {
	return encodeString(s, AlgorithmSHA256)
}

func SHA256File(path string) (string, error) {
	return encodeFile(path, AlgorithmSHA256)
}
