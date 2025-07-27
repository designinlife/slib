package crypto

func SHA512Bytes(b []byte) string {
	return encodeBytes(b, AlgorithmSHA512)
}

func SHA512String(s string) string {
	return encodeString(s, AlgorithmSHA512)
}

func SHA512File(path string) (string, error) {
	return encodeFile(path, AlgorithmSHA512)
}
