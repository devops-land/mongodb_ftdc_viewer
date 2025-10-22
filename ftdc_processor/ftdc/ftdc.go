package ftdc

import (
	"context"
)

// StreamBatches is a public wrapper to stream ftdc metrics
func StreamBatches(ctx context.Context, path string, metricsIncludeFilePath string, batchSize, buffer int) (<-chan StreamBatch, <-chan error) {
	return streamFTDCMetricsInBatches(ctx, path, metricsIncludeFilePath, batchSize, buffer)
}

func ReadMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	return readMetadata(ctx, path)
}
