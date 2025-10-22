package ftdc

import (
	"context"
	"github.com/evergreen-ci/birch"
	"github.com/jodevsa/ftdc"
)

type FTDCDataIterator struct {
	ctx      context.Context
	it       ftdc.Iterator
	doc      *birch.Document
	metadata *birch.Document
}

type StreamBatch struct {
	Items []map[string]interface{}
}
