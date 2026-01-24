package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func LoadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
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

func LoadPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	var publicKey *rsa.PublicKey
	if publicKeyPath == "" {
		publicKey = nil
	} else {
		publicKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			publicKey = nil
			fmt.Printf("couldn't read rsa public key from file: %s", publicKeyPath)
			return publicKey, err
		} else {
			publicKeyPemBlock, _ := pem.Decode(publicKeyBytes)
			if publicKeyPemBlock == nil {
				publicKey = nil
				log.Printf("warning: failed to decode PEM block from file: %s", publicKeyPath)
				return publicKey, fmt.Errorf("failed to decode PEM block from file: %s", publicKeyPath)
			} else {
				parsedKey, err := x509.ParsePKIXPublicKey(publicKeyPemBlock.Bytes)
				if err != nil {
					publicKey = nil
					log.Printf("warning: failed to parse PEM block from file: %s", publicKeyPath)
					return publicKey, err
				} else {
					var ok bool
					publicKey, ok = parsedKey.(*rsa.PublicKey)
					if !ok {
						publicKey = nil
						log.Printf("warning: key is not RSA public key")
						return publicKey, fmt.Errorf("key is not RSA public key")
					}
				}
			}
		}
	}

	return publicKey, nil
}
