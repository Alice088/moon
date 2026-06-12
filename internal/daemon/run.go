package daemon

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"moon/internal/analyzer"
	"moon/internal/collector"
	"moon/internal/config"
	"moon/internal/dispatcher"
	"moon/internal/entity"
	"moon/internal/hook"
	"moon/internal/notifier"
)

func Run(cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
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

	hks := []hook.Hook{
		hook.WriteAlertToDB(cfg.Storage.DBPath),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pip := entity.NewPipeline(collectors)
	pip.Interval = interval
	pipOut, pipErrs := pip.Run(ctx)

	anaPool := entity.NewAnalyzerPool(analyzers, cfg.AnalyzerWorkers)
	anaErrs, processed := anaPool.Run(ctx, pipOut)

	hookRunner := hook.NewRunner(hks, cfg.HookWorkers)
	hookErrs := hookRunner.Run(ctx, processed)

	dispErrs := dispatcher.Run(ctx, processed, notifiers)

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
