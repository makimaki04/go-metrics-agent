package agent

import (
	"math/rand"
	"runtime"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
)

type Collector struct {
	storage   Storage
	pollCount int64
}

func NewCollector(storage Storage) *Collector {
	return &Collector{storage: storage}
}

func (c *Collector) CollcetMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": float64(m.GCCPUFraction),
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
	}

	for name, val := range metrics {
		v := val
		c.storage.SetMetric(name, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		})
	}

	c.pollCount++
	c.storage.SetMetric("PollCount", models.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &c.pollCount,
	})

	randomValue := rand.Float64()
	c.storage.SetMetric("RandomValue", models.Metrics{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &randomValue,
	})

}
