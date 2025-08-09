package agent

import (
	"bytes"
	"compress/gzip"
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
}

type SenderStorageIntreface interface {
	GetAll() map[string]models.Metrics
}

func NewSender(client *resty.Client, url string, storage SenderStorageIntreface) *Sender {
	return &Sender{client: client, baseURL: url, storage: storage}
}

//Новая реализация отправки метрки к эндпоинту /update
func (s Sender) SendMetricsV2() error {
	url := fmt.Sprintf("%s/update", s.baseURL)
	metrics := s.storage.GetAll()

	for _, m := range metrics {
		var buf bytes.Buffer
		resp, err := json.MarshalIndent(m, "", "	")
		if err != nil {
			return fmt.Errorf("json serialize error")
		}

		w := gzip.NewWriter(&buf)
		w.Write(resp)
		if err := w.Close(); err != nil {
			return fmt.Errorf("gzip close error: %v", err)
		}

		response, err := s.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(buf.Bytes()).
			Post(url)
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
