package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"
)

func CryptoMiddleware(privateKey *rsa.PrivateKey, next http.HandlerFunc) http.HandlerFunc {
	if privateKey == nil {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		encryptedData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if len(encryptedData) == 0 {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		blockSize := privateKey.Size()
		var decryptedData []byte

		if len(encryptedData)%blockSize != 0 {
			http.Error(w, "Invalid encrypted data size", http.StatusBadRequest)
			return
		}

		for i := 0; i < len(encryptedData); i += blockSize {
			block := encryptedData[i : i+blockSize]
			decryptedBlock, err := rsa.DecryptOAEP(
				sha256.New(),
				rand.Reader,
				privateKey,
				block,
				nil,
			)
			if err != nil {
				http.Error(w, "Decryption failed", http.StatusBadRequest)
				return
			}
			decryptedData = append(decryptedData, decryptedBlock...)
		}

		r.Body = io.NopCloser(bytes.NewReader(decryptedData))
		r.ContentLength = int64(len(decryptedData))

		next(w, r)
	}
}
