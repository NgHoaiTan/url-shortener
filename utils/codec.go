package utils

import "crypto/rand"

const base62Chars = "QW8eRTYUIOPmNcpyVtBoSrEwixL5X1M3n6b9DAuvqC7z0Za2Ksd4JfgHhjGklF"

func GenerateShortCode(length int) (string, error) {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		b[i] = base62Chars[int(b[i])%len(base62Chars)]
	}

	return string(b), nil
}
