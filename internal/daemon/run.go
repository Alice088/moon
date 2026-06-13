package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"moon/internal/analyzer"
	"moon/internal/bot"
	"moon/internal/collector"
	"moon/internal/config"
	"moon/internal/dispatcher"
	"moon/internal/storage"
	"moon/internal/entity"
	"moon/internal/hook"
	"moon/internal/notifier"
)

func fanout(ctx context.Context, input <-chan *entity.Metrics) (<-chan *entity.Metrics, <-chan *entity.Metrics) {
	a := make(chan *entity.Metrics, 100)
	b := make(chan *entity.Metrics, 100)
	go func() {
		defer close(a)
		defer close(b)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-input:
				if !ok {
					return
				}
				select {
				case a <- m:
				case <-ctx.Done():
					return
				}
				select {
				case b <- m:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return a, b
}

func Run(cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	// ensure database directory exists
	dbDir := filepath.Dir(cfg.Storage.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("create db dir %s: %w", dbDir, err)
	}

	if cfg.Debug {
		storage.Debug = true
		log.Println("[debug] debug mode enabled")
	}

	interval := time.Second
	if cfg.CollectInterval != "" {
		if d, err := time.ParseDuration(cfg.CollectInterval); err == nil {
			interval = d
		}
	}

	collectors := []entity.Collector{
		collector.NewCPUCollector(),
		collector.NewRAMCollector(),
		collector.NewDiskCollector(),
	}

	analyzers := []entity.Analyzer{
		analyzer.NewCPUPeak(cfg.CPU.PeakThresholdPct),
		analyzer.NewRAMPeak(cfg.RAM.PeakThresholdPct),
		analyzer.NewDiskPeak(cfg.Disk.PeakThresholdPct),
	}

	notifiers := notifier.NewNotifiers(cfg.Notify)

	var botToken string
	var chatIDs []string
	for _, n := range cfg.Notify {
		if n.Type == "telegram" {
			if n.BotToken != "" && botToken == "" {
				botToken = n.BotToken
			}
			if n.ChatID != "" {
				chatIDs = append(chatIDs, n.ChatID)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if botToken != "" {
		b := bot.New(botToken, cfg.Storage.DBPath, chatIDs...)
		b.SetDebug(cfg.Debug)
		go b.Run(ctx)
	}

	hks := []hook.Hook{
		hook.WriteAlertToDB(cfg.Storage.DBPath),
	}

	pip := entity.NewPipeline(collectors)
	pip.Interval = interval
	pipOut, pipErrs := pip.Run(ctx)

	anaPool := entity.NewAnalyzerPool(analyzers, cfg.AnalyzerWorkers)
	anaErrs, processed := anaPool.Run(ctx, pipOut)

	hookCh, dispCh := fanout(ctx, processed)

	hookRunner := hook.NewRunner(hks, cfg.HookWorkers)
	hookErrs := hookRunner.Run(ctx, hookCh)

	dispErrs := dispatcher.Run(ctx, dispCh, notifiers)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			log.Println("shutting down")
			cancel()
			time.Sleep(1 * time.Second)
			return nil

		case err := <-pipErrs:
			if err != nil {
				log.Printf("pipeline error: %v", err)
			}
		case err := <-anaErrs:
			if err != nil {
				log.Printf("analyzer error: %v", err)
			}
		case err := <-hookErrs:
			if err != nil {
				log.Printf("hook error: %v", err)
			}
		case err := <-dispErrs:
			if err != nil {
				log.Printf("dispatcher error: %v", err)
			}

		case <-ticker.C:
			log.Println("collecting...")
		}
	}
}
