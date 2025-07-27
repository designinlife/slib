package crypto

func MD5Bytes(b []byte) string {
	return encodeBytes(b, AlgorithmMD5)
}

func MD5String(s string) string {
	return encodeString(s, AlgorithmMD5)
}

func MD5File(path string) (string, error) {
	return encodeFile(path, AlgorithmMD5)
}
