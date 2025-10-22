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
	"path/filepath"
	"sync/atomic"
	"time"
)

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

	logging.Info("Processing: %s", absInputPath)
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

	logging.Info("Completed processing %s", absInputPath)

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
				fmt.Printf("\rIngested %-20s of diagnostics metrics", duration)
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

	logging.Info("finished process all the %d files!", len(files))
}
