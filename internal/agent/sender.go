package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

type Sender struct {
	client  *resty.Client
	baseURL string
	storage SenderStorageIntreface
	key []byte
}

type SenderStorageIntreface interface {
	GetAll() map[string]models.Metrics
}

func NewSender(client *resty.Client, url string, storage SenderStorageIntreface, key string) *Sender {
	return &Sender{
		client: client, 
		baseURL: url, 
		storage: storage, 
		key: []byte(key),
	}
}

func prepareGzipBody(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	resp, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("json serialize error")
	}

	w := gzip.NewWriter(&buf)
	if _, err := w.Write(resp); err != nil {
    	return nil, fmt.Errorf("gzip write error: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("gzip close error: %v", err)
	}

	return buf.Bytes(), nil
}

//Новая реализация отправки метрки к эндпоинту /update
func (s Sender) SendMetricsV2() error {
	url := fmt.Sprintf("%s/update", s.baseURL)
	metrics := s.storage.GetAll()

	for _, m := range metrics {
		body, err := prepareGzipBody(m)
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

func (s *Sender) sendBatch(url string, batch []models.Metrics) error {
	body, err := prepareGzipBody(batch)
	if err != nil {
		return err
	}

	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body)

	if len(s.key) > 0 {
		json, err := json.Marshal(batch)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(append(json, s.key...))
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

// Старая реализация к эндпоинту update/{MType}/{ID}/{value}
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
