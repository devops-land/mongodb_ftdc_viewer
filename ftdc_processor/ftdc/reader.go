package ftdc

import (
	"context"
	"github.com/mongodb/ftdc"
	"io"
	"os"
)

func readMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return map[string]interface{}{}, err
	}

	cs := ftdc.ReadChunks(ctx, file)
	metadata := make(map[string]interface{})
	for cs.Next() {
		md := cs.Chunk().GetMetadata()
		if md != nil {
			metadata = normalizeDocument(md, map[string]struct{}{})
			break
		}
	}
	return metadata, nil
}

func readFTDCData(ctx context.Context, r io.Reader) *FTDCDataIterator {

	iter := &FTDCDataIterator{
		ctx: ctx,
		it:  ftdc.ReadMetrics(ctx, r),
	}

	print(iter.it.Metadata().String())

	return iter
}
