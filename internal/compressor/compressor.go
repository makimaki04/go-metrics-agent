package compressor

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// prepareGzipBody - method for preparing the gzip body
// prepare the gzip body
// if error, return error
// if success, return the gzip body
func PrepareEncryptedGzipBody(data interface{}, publicKey *rsa.PublicKey) ([]byte, error) {
	resp, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("json serialize error")
	}

	if publicKey != nil {
		maxBlockSize := publicKey.Size() - 2*sha256.Size - 2
		var encryptedData []byte

		for i := 0; i < len(resp); i += maxBlockSize {
			end := i + maxBlockSize
			if end > len(resp) {
				end = len(resp)
			}

			block := resp[i:end]
			encryptedBlock, err := rsa.EncryptOAEP(
				sha256.New(),
				rand.Reader,
				publicKey,
				block,
				nil,
			)
			if err != nil {
				return nil, fmt.Errorf("rsa encryption error: %v", err)
			}
			encryptedData = append(encryptedData, encryptedBlock...)
		}

		gzipData, err := PrepareGzipBody(encryptedData)
		if err != nil {
			return nil, err
		}
		return gzipData, nil
	}

	gzipData, err := PrepareGzipBody(resp)
	if err != nil {
		return nil, err
	}
	return gzipData, nil
}

func PrepareGzipBody(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write error: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("gzip close error: %v", err)
	}

	return buf.Bytes(), nil
}