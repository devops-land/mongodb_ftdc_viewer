package main

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/yourusername/my-ftdc-tool/ftdc"
	"github.com/yourusername/my-ftdc-tool/internal/config"
	"github.com/yourusername/my-ftdc-tool/internal/logging"
	"golang.org/x/sync/errgroup"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"sync/atomic"
	"time"
)

func getFTDCFileInfluxTags(ctx context.Context, path string) (map[string]string, error) {
	metadata, err := ftdc.ReadMetadata(ctx, path)

	if err != nil {
		return map[string]string{}, err
	}

	hostname := metadata["doc"].(map[string]interface{})["hostInfo"].(map[string]interface{})["system"].(map[string]interface{})["hostname"].(string)
	version := metadata["doc"].(map[string]interface{})["buildInfo"].(map[string]interface{})["version"].(string)

	return map[string]string{
		"hostname": hostname,
		"version":  version,
	}, nil
}

func ingestFTDCFromFile(absInputPath string, cfg *config.Config, counter *atomic.Int64) error {
	client := influxdb2.NewClientWithOptions(cfg.InfluxURL, cfg.InfluxToken, influxdb2.DefaultOptions().SetUseGZip(cfg.InfluxUseGZip).SetPrecision(time.Second).SetMaxRetries(5).SetMaxRetryInterval(10000))
	defer client.Close()

	w := client.WriteAPIBlocking(cfg.InfluxOrg, cfg.InfluxBucket)
	ctx := context.Background()

	tags, err := getFTDCFileInfluxTags(ctx, absInputPath)
	if err != nil {
		return err
	}

	batches, errs := ftdc.StreamBatches(ctx, absInputPath, cfg.MetricsIncludeFile, cfg.BatchSize, cfg.BatchBuffer)
	total := 0

	logging.Info("Processing: %s", absInputPath)
	for batch := range batches {
		points := []*write.Point{}

		for _, doc := range batch.Items {
			t := time.UnixMilli(doc["start"].(int64))
			points = append(points, influxdb2.NewPoint("ftdc", tags, doc, t))
		}

		if err := w.WritePoint(context.Background(), points...); err != nil {
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
