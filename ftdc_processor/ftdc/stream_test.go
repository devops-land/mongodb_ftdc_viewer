package ftdc

import (
	"os"
	"testing"
)

func BenchmarkStreamFTDCMetricsInBatches(b *testing.B) {
	const (
		path        = "./testdata/metrics.2025-10-05T13-40-00Z-00000"
		includeFile = "./testdata/metrics_to_get.txt"
	)

	fi, err := os.Stat(path)
	if err != nil {
		b.Fatalf("stat %s: %v", path, err)
	}
	fileSize := fi.Size()

	cases := []struct {
		name       string
		batchSize  int
		bufferSize int
	}{
		{"batch1k_workers3", 20, 3},
		{"batch1k_workers3", 20, 10},
		{"batch1k_workers3", 20, 20},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(fileSize) // bytes processed per iteration

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				batches, _ := streamFTDCMetricsInBatches(b.Context(), path, includeFile, tc.batchSize, tc.bufferSize)

				count := 0
				for range batches {
					count++
				}
				// Prevent compiler from optimizing away the loop.
				if count == 0 {
					b.Fatalf("no batches produced (batchSize=%d, workers=%d)", tc.batchSize, tc.bufferSize)
				}
			}
		})
	}
}
