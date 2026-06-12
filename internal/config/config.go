package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"moon/pkg/types"
)

type Config struct {
	CPU             CPUConfig `yaml:"cpu"`
	AnalyzerWorkers int       `yaml:"analyzer_workers"`
}

type CPUConfig struct {
	PeakThresholdPct types.Percent `yaml:"peak_threshold_pct"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return &cfg, nil
}
