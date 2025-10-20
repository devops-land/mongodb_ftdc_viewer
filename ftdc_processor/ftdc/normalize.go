package ftdc

import (
	"github.com/evergreen-ci/birch"
	"github.com/evergreen-ci/birch/bsontype"
	"regexp"
	"time"
)

var diskMetricsPattern = regexp.MustCompile(`^systemMetrics\.disks\..*\.(io_in_progress|io_queued_ms|io_time_ms|read_sectors|read_time_ms|reads|reads_merged|write_sectors|write_time_ms|writes|writes_merged)$`)
var mountMetricsPattern = regexp.MustCompile(`^systemMetrics\.mounts\.(\/(?:[^\/]+\/?)*)\.(available|capacity|free)$`)

func isIncluded(key string, includedPatterns []string) bool {

	if len(includedPatterns) == 0 {
		return true
	}

	if key == "start" {
		return true
	}

	for _, pattern := range includedPatterns {
		if pattern == key {
			return true
		}
	}

	if diskMetricsPattern.MatchString(key) {
		return true
	}
	if mountMetricsPattern.MatchString(key) {
		return true
	}
	return false
}

func normalizeDocument(document *birch.Document, includedPatterns []string) map[string]interface{} {
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

func normalizeValue(val *birch.Value, includedPatterns []string) interface{} {
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
