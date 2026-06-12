package entity

import (
	"sync"
	"time"
)

type Metrics struct {
	Mu        sync.RWMutex   `json:"mu"`
	Data      map[string]any `json:"data"`
	Timestamp time.Time      `json:"timestamp"`
}

func NewMetrics(name string) *Metrics {
	return &Metrics{
		Data:      make(map[string]any),
		Timestamp: time.Now(),
	}
}

func (m *Metrics) Set(key string, value any) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Data[key] = value
}

func (m *Metrics) Get(key string) any {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	return m.Data[key]
}

func (m *Metrics) GetAll() map[string]any {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	copy := make(map[string]any)
	for k, v := range m.Data {
		copy[k] = v
	}
	return copy
}
