package crypto

func SHA384Bytes(b []byte) string {
	return encodeBytes(b, AlgorithmSHA384)
}

func SHA384String(s string) string {
	return encodeString(s, AlgorithmSHA384)
}

func SHA384File(path string) (string, error) {
	return encodeFile(path, AlgorithmSHA384)
}
