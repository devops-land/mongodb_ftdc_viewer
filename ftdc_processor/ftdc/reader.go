package ftdc

import (
	"context"
	"github.com/jodevsa/ftdc"
	"io"
)

func readFTDCData(ctx context.Context, r io.Reader) *FTDCDataIterator {

	iter := &FTDCDataIterator{
		ctx: ctx,
		it:  ftdc.ReadMetrics(ctx, r),
	}

	print(iter.it.Metadata().String())

	return iter
}
