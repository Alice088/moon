package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"moon/pkg/types"
)

type Config struct {
	CPU             CPUConfig      `yaml:"cpu"`
	RAM             RAMConfig      `yaml:"ram"`
	Disk            DiskConfig     `yaml:"disk"`
	AnalyzerWorkers int            `yaml:"analyzer_workers"`
	Notify          []NotifyConfig `yaml:"notify"`
}

type CPUConfig struct {
	PeakThresholdPct types.Percent `yaml:"peak_threshold_pct"`
}

type RAMConfig struct {
	PeakThresholdPct types.Percent `yaml:"peak_threshold_pct"`
}

type DiskConfig struct {
	PeakThresholdPct types.Percent `yaml:"peak_threshold_pct"`
}

type NotifyConfig struct {
	Type      string `yaml:"type"`
	BotToken  string `yaml:"bot_token"`
	SmtpHost  string `yaml:"smtp_host"`
	SmtpPort  int    `yaml:"smtp_port"`
	SmtpUser  string `yaml:"smtp_user"`
	SmtpPass  string `yaml:"smtp_pass"`
	FromAddr  string `yaml:"from_addr"`
	ToAddr    string `yaml:"to_addr"`
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
