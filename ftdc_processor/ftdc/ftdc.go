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

func GetTags(ctx context.Context, path string) (map[string]string, error) {
	metadata, err := ReadMetadata(ctx, path)
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
