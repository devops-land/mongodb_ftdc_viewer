package ftdc

import (
	"github.com/evergreen-ci/birch"
	"github.com/evergreen-ci/birch/bsontype"
	"strings"
	"time"
)

func isMountMetric(key string) bool {
	var p = "systemMetrics.mounts."
	if !strings.HasPrefix(key, p) {
		return false
	}
	for _, field := range []string{
		"available", "capacity", "free",
	} {
		if strings.Contains(key, field) {
			return true
		}
	}
	return false
}

func isDiskMetric(key string) bool {
	var p = "systemMetrics.disks."
	if !strings.HasPrefix(key, p) {
		return false
	}
	for _, field := range []string{
		"io_in_progress", "io_queued_ms", "io_time_ms",
		"read_sectors", "read_time_ms", "reads", "reads_merged",
		"write_sectors", "write_time_ms", "writes", "writes_merged",
	} {
		if strings.Contains(key, field) {
			return true
		}
	}
	return false
}

func isIncluded(key string, includedPatterns map[string]struct{}) bool {

	if len(includedPatterns) == 0 {
		return true
	}

	if key == "start" {
		return true
	}

	if _, ok := includedPatterns[key]; ok {
		return true
	}

	if isDiskMetric(key) || isMountMetric(key) {
		return true
	}

	return false
}

func normalizeDocument(document *birch.Document, includedPatterns map[string]struct{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	iter := document.Iterator()

	for iter.Next() {
		elem := iter.Element()
		key := elem.Key()
		val := elem.Value()
		if isIncluded(key, includedPatterns) {
			normalized[key] = normalizeValue(val, includedPatterns)
		}
	}
	return normalized
}

func normalizeValue(val *birch.Value, includedPatterns map[string]struct{}) interface{} {
	switch val.Type() {
	case bsontype.Double:
		return val.Double()
	case bsontype.String:
		return val.StringValue()
	case bsontype.EmbeddedDocument:
		return normalizeDocument(val.MutableDocument(), includedPatterns)
	case bsontype.Boolean:
		return val.Boolean()
	case bsontype.Int32:
		return val.Int32()
	case bsontype.Int64:
		return val.Int64()
	case bsontype.Null:
		return -1
	case bsontype.ObjectID:
		return val.ObjectID().Hex()
	case bsontype.Array:
		out := []interface{}{}
		it := val.MutableArray().Iterator()
		i := 0
		for it.Next() {
			out = append(out, normalizeValue(it.Value(), includedPatterns))
			i++
		}
		return out
	case bsontype.DateTime:
		return time.UnixMilli(val.DateTime()).UnixMilli()
	default:
		// Handle unsupported types as raw or string, or skip
		return val.Interface() // fallback
	}
}
