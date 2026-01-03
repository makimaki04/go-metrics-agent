package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func LoadCryptoKey(keyPath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read file error: %v", err)
	}

	keyPemBlock, _ := pem.Decode(keyBytes)
	if keyPemBlock == nil {
		return nil, fmt.Errorf("decode pem block error")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(keyPemBlock.Bytes)
	if err != nil {
		privateKey, err := x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("pem block parse error: %v", err)
		}
		return privateKey, nil
	}

	rsaPrivateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("couldn't assert rsa.PrivateKey type")
	}

	return rsaPrivateKey, nil
}