package main

import (
	"context"
	"fmt"
	"github.com/yourusername/my-ftdc-tool/ftdc"
	"github.com/yourusername/my-ftdc-tool/internal/config"
	"github.com/yourusername/my-ftdc-tool/internal/influx"
	"github.com/yourusername/my-ftdc-tool/internal/logging"
	"golang.org/x/sync/errgroup"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync/atomic"
	"syscall"
	"time"
)

func buildGrafanaURL(ctx context.Context, cfg *config.Config) (error, string) {
	client := influx.NewClient(ctx, influx.Config{
		Org:         cfg.InfluxOrg,
		Bucket:      cfg.InfluxBucket,
		Url:         cfg.InfluxURL,
		Token:       cfg.InfluxToken,
		UseGzip:     cfg.InfluxUseGZip,
		Measurement: cfg.InfluxMeasurement,
	})
	defer client.Close()
	err, earliest := client.FetchEarliestTimestamp()
	if err != nil {
		return err, ""
	}

	err, latest := client.FetchLatestTimestamp()
	if err != nil {
		return err, ""
	}
	baseURL := "http://localhost:3001/d/ddnw277huiv40ae/ftdc-dashboard"
	return nil, fmt.Sprintf("%s?from=%s&to=%s&timezone=UTC", baseURL, earliest, latest)
}

func ingestFTDCFromFile(absInputPath string, cfg *config.Config, counter *atomic.Int64) error {
	ctx := context.Background()
	client := influx.NewClient(ctx, influx.Config{
		Org:         cfg.InfluxOrg,
		Bucket:      cfg.InfluxBucket,
		Url:         cfg.InfluxURL,
		Token:       cfg.InfluxToken,
		UseGzip:     cfg.InfluxUseGZip,
		Measurement: cfg.InfluxMeasurement,
	})
	defer client.Close()

	tags, err := ftdc.GetTags(ctx, absInputPath)
	if err != nil {
		return err
	}

	batches, errs := ftdc.StreamBatches(ctx, absInputPath, cfg.MetricsIncludeFile, cfg.BatchSize, cfg.BatchBuffer)
	total := 0
	if cfg.Debug {
		logging.Info("Processing: %s", absInputPath)
	}
	for batch := range batches {
		var points []*influx.Point
		for _, doc := range batch.Items {
			t := time.UnixMilli(doc["start"].(int64))
			points = append(points, client.NewPoint(tags, doc, t))
		}

		if err := client.WritePoint(points...); err != nil {
			return err
		}

		total += len(batch.Items)
		counter.Add(int64(len(batch.Items)))

	}

	// 5. check for stream errors
	if err := <-errs; err != nil && err != io.EOF {
		fmt.Println("stream error:", err)
	}
	if cfg.Debug {
		logging.Info("Completed processing %s", absInputPath)
	}
	return nil
}

func main() {
	cfg := config.ParseFlags()

	var processed atomic.Int64

	time.Sleep(5 * time.Second)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				duration := time.Duration(processed.Load()) * time.Second
				timestamp := time.Now().Format("15:04:05")
				fmt.Printf("\r[%s] Ingested %-15s of diagnostics metrics", timestamp, duration)
			case <-done:
				return
			}
		}
	}()

	logging.PrintBanner()
	cfg.Print()
	// Ensure output file path is absolute
	absFTDCDirectory, err := filepath.Abs(cfg.InputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path of output file: %v", err)
	}
	var files []string
	err = filepath.WalkDir(absFTDCDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(cfg.Parallel)
	// sort files in ascending order
	sort.Strings(files)

	logging.Info("%d files queued for processing", len(files))
	for _, f := range files {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return ingestFTDCFromFile(filepath.Clean(f), cfg, &processed)
			}
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("failed:", err)
	}

	// stop the periodic log updates
	close(done)

	err, grafanaUrl := buildGrafanaURL(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	logging.Info("Metrics available for analysis on:\n\n %s\n", grafanaUrl)
	if cfg.WaitForever {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Println("Press Ctrl+C to exit.")
		<-ctx.Done()
		fmt.Println("Received shutdown signal.")
	}
}
