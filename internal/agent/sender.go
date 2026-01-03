package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

// Sender - struct for the sender
type Sender struct {
	client    *resty.Client
	baseURL   string
	storage   SenderStorageIntreface
	key       []byte
	publicKey *rsa.PublicKey
}

// SenderStorageIntreface - interface for the sender storage
// GetAll - method for getting all metrics from the storage
type SenderStorageIntreface interface {
	GetAll() map[string]models.Metrics
}

// NewSender - method for creating a new sender
// create a new sender
// if success, return nil
func NewSender(client *resty.Client, url string, storage SenderStorageIntreface, key string, publicKeyPath string) *Sender {
	var publicKey *rsa.PublicKey

	if publicKeyPath == "" {
		publicKey = nil
	} else {
		publicKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			fmt.Printf("couldn't read rsa public key from file: %s, continuing without encryption", publicKeyPath)
			publicKey = nil
		} else {
			publicKeyPemBlock, _ := pem.Decode(publicKeyBytes)
			if publicKeyPemBlock == nil {
				publicKey = nil
				log.Printf("warning: failed to decode PEM block from file: %s, continuing without encryption", publicKeyPath)
			} else {
				parsedKey, err := x509.ParsePKIXPublicKey(publicKeyPemBlock.Bytes)
				if err != nil {
					publicKey = nil
					log.Printf("warning: failed to parse PEM block from file: %s, continuing without encryption", publicKeyPath)
				} else {
					var ok bool
					publicKey, ok = parsedKey.(*rsa.PublicKey)
					if !ok {
						publicKey = nil
						log.Printf("warning: key is not RSA public key, continuing without encryption")
					}
				}
			}
		}
	}

	return &Sender{
		client:    client,
		baseURL:   url,
		storage:   storage,
		key:       []byte(key),
		publicKey: publicKey,
	}
}

// prepareGzipBody - method for preparing the gzip body
// prepare the gzip body
// if error, return error
// if success, return the gzip body
func prepareEncryptedGzipBody(data interface{}, publicKey *rsa.PublicKey) ([]byte, error) {
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

		gzipData, err := prepareGzipBody(encryptedData)
		if err != nil {
			return nil, err
		}
		return gzipData, nil
	}

	gzipData, err := prepareGzipBody(resp)
	if err != nil {
		return nil, err
	}
	return gzipData, nil
}

func prepareGzipBody(data []byte) ([]byte, error) {
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

// SendMetricsV2 - method for sending metrics to the server
// send the metrics to the server
// if error, return error
// if success, return nil
func (s Sender) SendMetricsV2() error {
	url := fmt.Sprintf("%s/update", s.baseURL)
	metrics := s.storage.GetAll()

	for _, m := range metrics {
		body, err := prepareEncryptedGzipBody(m, s.publicKey)
		if err != nil {
			return err
		}

		req := s.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(body)

		if len(s.key) > 0 {
			hash := sha256.Sum256(append(body, s.key...))
			hex := hex.EncodeToString(hash[:])
			req.SetHeader("HashSHA256", hex)
		}

		response, err := req.Post(url)
		if err != nil {
			log.Printf("failed to send metric %s: %v", m.ID, err)
			return err
		}

		if response.StatusCode() != http.StatusOK {
			return fmt.Errorf("something went wrong. bad status: %s", response.Status())
		}
		log.Printf("Sending %s %s", m.MType, m.ID)
		log.Printf("%v", response.Status())
	}
	return nil
}

// SendMetricsBatch - method for sending metrics batch to the server
// send the metrics batch to the server
// if error, return error
// if success, return nil
func (s Sender) SendMetricsBatch(batch []models.Metrics) error {
	url := fmt.Sprintf("%s/updates", s.baseURL)
	metrics := s.storage.GetAll()
	batchCopy := batch

	for _, m := range metrics {
		batchCopy = append(batchCopy, m)
		if len(batchCopy) == 100 {
			if err := s.sendBatch(url, batchCopy); err != nil {
				return err
			}
			batchCopy = batchCopy[:0]
		}
	}

	if len(batchCopy) > 0 {
		if err := s.sendBatch(url, batchCopy); err != nil {
			return err
		}
	}

	return nil
}

// sendBatch - method for sending a batch of metrics to the server
// send the batch of metrics to the server
// if error, return error
// if success, return nil
func (s *Sender) sendBatch(url string, batch []models.Metrics) error {
	body, err := prepareEncryptedGzipBody(batch, s.publicKey)
	if err != nil {
		return err
	}

	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body)

	if len(s.key) > 0 {
		hash := sha256.Sum256(append(body, s.key...))
		hex := hex.EncodeToString(hash[:])
		req.SetHeader("HashSHA256", hex)
	}

	response, err := req.Post(url)
	if err != nil {
		log.Printf("failed to send metric batch: %v", err)
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("something went wrong. bad status: %s", response.Status())
	}

	log.Printf("Sending batch %v", len(batch))
	log.Printf("%v", response.Status())

	return nil
}

// old realization of sending metrics to the server
// send the metrics to the server
// if error, return error
// if success, return nil
func (s Sender) SendMetrics() error {
	metrics := s.storage.GetAll()
	for _, m := range metrics {
		var value string
		switch m.MType {
		case "gauge":
			value = fmt.Sprintf("%v", *m.Value)
		case "counter":
			value = fmt.Sprintf("%d", *m.Delta)
		default:
			continue
		}

		url := fmt.Sprintf("%s/update/%s/%s/%s", s.baseURL, m.MType, m.ID, value)

		response, err := s.client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)
		if err != nil {
			log.Printf("failed to send metric %s: %v", m.ID, err)
			return err
		}

		if response.StatusCode() != http.StatusOK {
			return fmt.Errorf("something went wrong. bad status: %s", response.Status())
		}
		log.Printf("Sending %s %s = %s", m.MType, m.ID, value)
		log.Printf("%v", response.Status())
	}
	return nil
}
