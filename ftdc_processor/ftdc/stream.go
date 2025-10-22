package ftdc

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func streamFTDCMetricsInBatches(ctx context.Context, path string, metricsIncludeFilePath string, batchSize, buffer int) (<-chan StreamBatch, <-chan error) {
	metricsIncludeFile, err := os.Open(metricsIncludeFilePath)
	if err != nil {
		fmt.Errorf("couldn't open BSON file: %v", err)

	}

	defer metricsIncludeFile.Close()

	file, err := os.Open(path)
	if err != nil {
		fmt.Errorf("couldn't open BSON file: %v", err)

	}

	scanner := bufio.NewScanner(metricsIncludeFile)
	includePatterns := make(map[string]struct{})
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			includePatterns[line] = struct{}{}
		}
	}

	out := make(chan StreamBatch, buffer)
	errc := make(chan error, 1)

	iter := readFTDCData(ctx, file)

	go func() {
		defer close(out)
		defer close(errc)

		for {
			sb := StreamBatch{
				Items: make([]map[string]interface{}, 0, batchSize),
			}

			for i := 0; i < batchSize; i++ {
				if iter.Next() {
					sb.Items = append(sb.Items, iter.NormalisedDocument(includePatterns))
				} else {
					break
				}
			}
			if len(sb.Items) == 0 {
				return
			}
			select {
			case out <- sb:
			case <-ctx.Done():
				errc <- ctx.Err()
				return
			}
		}
	}()

	return out, errc
}
