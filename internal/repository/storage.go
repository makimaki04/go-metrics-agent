package repository

type MemStorage struct {
	gauges map[string]float64
	counters map[string]int64
}

type Storage interface {
	setGauge(name string, value float64) error
	getGauge(name string)

	setCounter(name string, value int64) error
	getCounter(name string)
}


