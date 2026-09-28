package ftdc

import (
	"strconv"

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

func joinMetricKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// isTopLevelRoleWrapper reports MongoDB 8+ FTDC role document keys that must not
// appear in the flattened metric path. Only matches at the metrics document root
// so nested fields named "shard" on older versions are left alone.
func isTopLevelRoleWrapper(prefix, key string) bool {
	if prefix != "" {
		return false
	}
	switch key {
	case "shard", "router", "common":
		return true
	default:
		return false
	}
}

// stripRoleKeyPrefix removes a leading role prefix from flattened metric keys.
// github.com/mongodb/ftdc often emits already-dotted keys such as
// "shard.replSetGetStatus.members.0.state" at the document root.
// Nested keys (prefix != "") are left unchanged except for the historical
// "common." trim applied by the caller.
func stripRoleKeyPrefix(prefix, key string) string {
	if prefix != "" {
		return key
	}
	for _, p := range []string{"shard.", "router.", "common."} {
		if strings.HasPrefix(key, p) {
			return strings.TrimPrefix(key, p)
		}
	}
	return key
}

func setIfAbsent(dst map[string]interface{}, key string, value interface{}) {
	if _, exists := dst[key]; exists {
		return
	}
	dst[key] = value
}

// normalizeDocument preserves nested maps (used for FTDC metadata).
func normalizeDocument(document *birch.Document, includedPatterns map[string]struct{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	iter := document.Iterator()

	for iter.Next() {
		elem := iter.Element()
		key := elem.Key()
		key = strings.TrimPrefix(key, "common.")
		val := elem.Value()
		if isIncluded(key, includedPatterns) {
			setIfAbsent(normalized, key, normalizeValue(val, includedPatterns))
		}
	}
	return normalized
}

// normalizeMetricsDocument flattens nested FTDC samples into dotted field names for InfluxDB.
func normalizeMetricsDocument(document *birch.Document, includedPatterns map[string]struct{}) map[string]interface{} {
	return flattenDocument(document, "", includedPatterns)
}

func flattenDocument(document *birch.Document, prefix string, includedPatterns map[string]struct{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	iter := document.Iterator()

	for iter.Next() {
		elem := iter.Element()
		key := elem.Key()
		key = strings.TrimPrefix(key, "common.")
		key = stripRoleKeyPrefix(prefix, key)
		val := elem.Value()

		// MongoDB 8+ sharded FTDC wraps mongod/mongos collectors under shard/router/common.
		if isTopLevelRoleWrapper(prefix, key) && val.Type() == bsontype.EmbeddedDocument {
			for k, v := range flattenDocument(val.MutableDocument(), prefix, includedPatterns) {
				setIfAbsent(normalized, k, v)
			}
			continue
		}

		fullKey := joinMetricKey(prefix, key)

		switch val.Type() {
		case bsontype.EmbeddedDocument:
			for k, v := range flattenDocument(val.MutableDocument(), fullKey, includedPatterns) {
				setIfAbsent(normalized, k, v)
			}
		case bsontype.Array:
			flattenArray(val.MutableArray(), fullKey, includedPatterns, normalized)
		default:
			if isIncluded(fullKey, includedPatterns) {
				setIfAbsent(normalized, fullKey, normalizeScalar(val))
			}
		}
	}
	return normalized
}

func flattenArray(arr *birch.Array, prefix string, includedPatterns map[string]struct{}, out map[string]interface{}) {
	it := arr.Iterator()
	i := 0
	for it.Next() {
		indexKey := joinMetricKey(prefix, strconv.Itoa(i))
		val := it.Value()
		switch val.Type() {
		case bsontype.EmbeddedDocument:
			for k, v := range flattenDocument(val.MutableDocument(), indexKey, includedPatterns) {
				setIfAbsent(out, k, v)
			}
		case bsontype.Array:
			flattenArray(val.MutableArray(), indexKey, includedPatterns, out)
		default:
			if isIncluded(indexKey, includedPatterns) {
				setIfAbsent(out, indexKey, normalizeScalar(val))
			}
		}
		i++
	}
}

func normalizeScalar(val *birch.Value) interface{} {
	switch val.Type() {
	case bsontype.Double:
		return val.Double()
	case bsontype.String:
		return val.StringValue()
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
	case bsontype.DateTime:
		return time.UnixMilli(val.DateTime()).UnixMilli()
	default:
		return val.Interface()
	}
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
		for it.Next() {
			out = append(out, normalizeValue(it.Value(), includedPatterns))
		}
		return out
	case bsontype.DateTime:
		return time.UnixMilli(val.DateTime()).UnixMilli()
	default:
		return val.Interface()
	}
}
