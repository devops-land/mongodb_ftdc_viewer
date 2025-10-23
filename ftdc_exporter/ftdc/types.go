package ftdc

import (
	"context"
	"errors"
	"github.com/evergreen-ci/birch"
	"github.com/mongodb/ftdc"
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

var ErrInvalidFormat = errors.New("invalid ftdc format")
