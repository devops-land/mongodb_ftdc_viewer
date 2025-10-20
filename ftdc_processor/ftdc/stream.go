package ftdc

import (
	"bufio"
	"context"
	"fmt"
	"github.com/jodevsa/ftdc"
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
	var includePatterns []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			includePatterns = append(includePatterns, line)
		}
	}

	out := make(chan StreamBatch, buffer)
	errc := make(chan error, 1)

	iter := readFTDCData(ctx, file)

	file2, err := os.Open(path)
	if err != nil {
		fmt.Errorf("couldn't open BSON file: %v", err)

	}
	defer file2.Close()
	cs := ftdc.ReadChunks(ctx, file2)
	metadata := make(map[string]interface{})
	for cs.Next() {
		md := cs.Chunk().GetMetadata()
		if md != nil {
			metadata = normalizeDocument(md, []string{})
			break
		}
	}

	go func() {
		defer close(out)
		defer close(errc)

		for {
			sb := StreamBatch{
				Items:    make([]map[string]interface{}, 0, batchSize),
				Metadata: metadata,
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
