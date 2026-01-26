package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Ошибка генерации ключа: %v\n", err)
		os.Exit(1)
	}

	privateKeyFile, err := os.Create("private_key.pem")
	if err != nil {
		fmt.Printf("Ошибка создания файла: %v\n", err)
		os.Exit(1)
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		fmt.Printf("Ошибка записи приватного ключа: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Приватный ключ сохранен в private_key.pem")

	publicKey := &privateKey.PublicKey

	publicKeyFile, err := os.Create("public_key.pem")
	if err != nil {
		fmt.Printf("Ошибка создания файла: %v\n", err)
		os.Exit(1)
	}
	defer publicKeyFile.Close()

	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Printf("Ошибка маршалинга публичного ключа: %v\n", err)
		os.Exit(1)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}

	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		fmt.Printf("Ошибка записи публичного ключа: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Публичный ключ сохранен в public_key.pem")
	fmt.Println("\nКлючи успешно сгенерированы!")
	fmt.Println("Используйте:")
	fmt.Println("  - private_key.pem для сервера (флаг -crypto-key)")
	fmt.Println("  - public_key.pem для агента (флаг -crypto-key)")
}
