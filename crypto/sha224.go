package crypto

func SHA224Bytes(b []byte) string {
	return encodeBytes(b, AlgorithmSHA224)
}

func SHA224String(s string) string {
	return encodeString(s, AlgorithmSHA224)
}

func SHA224File(path string) (string, error) {
	return encodeFile(path, AlgorithmSHA224)
}
