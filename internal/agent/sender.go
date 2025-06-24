package agent

import (
	"fmt"
	"log"
	"net/http"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

type Sender struct {
	client  *http.Client
	baseUrl string
}

func NewSender(client *http.Client, url string) *Sender {
	return &Sender{client: client, baseUrl: url}
}

func (s Sender) SendMetrics(metrics map[string]models.Metrics) {
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

		url := fmt.Sprintf("%s/update/%s/%s/%s", s.baseUrl, m.MType, m.ID, value)

		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Printf("failed to create request for metric %s: %v", m.ID, err)
			continue
		}

		request.Header.Set("Content-Type", "text/plain")

		response, err := s.client.Do(request)
		if err != nil {
			log.Printf("failed to send metric %s: %v", m.ID, err)
			continue
		}
		log.Printf("Sending %s %s = %s", m.MType, m.ID, value)
		response.Body.Close()
	}
}
