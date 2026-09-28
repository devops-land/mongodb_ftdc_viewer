package ftdc

import (
	"testing"

	"github.com/evergreen-ci/birch"
)

func TestNormalizeMetricsDocument_flatAndNestedKeys(t *testing.T) {
	doc := birch.NewDocument(
		birch.EC.Int64("start", 1700000000000),
		birch.EC.Int64("serverStatus.wiredTiger.cache.bytes currently in the cache", 4096),
		birch.EC.SubDocument("serverStatus", birch.NewDocument(
			birch.EC.SubDocument("wiredTiger", birch.NewDocument(
				birch.EC.SubDocument("concurrentTransactions", birch.NewDocument(
					birch.EC.SubDocument("monitor", birch.NewDocument(
						birch.EC.Int32("available", 100),
						birch.EC.Int32("totalTickets", 128),
						birch.EC.Int32("queueLength", 0),
					)),
				)),
			)),
		)),
	)

	include := map[string]struct{}{
		"start": {},
		"serverStatus.wiredTiger.cache.bytes currently in the cache":          {},
		"serverStatus.wiredTiger.concurrentTransactions.monitor.available":      {},
		"serverStatus.wiredTiger.concurrentTransactions.monitor.totalTickets":   {},
		"serverStatus.wiredTiger.concurrentTransactions.monitor.queueLength":    {},
	}

	got := normalizeMetricsDocument(doc, include)

	want := map[string]interface{}{
		"start": int64(1700000000000),
		"serverStatus.wiredTiger.cache.bytes currently in the cache":          int64(4096),
		"serverStatus.wiredTiger.concurrentTransactions.monitor.available":    int32(100),
		"serverStatus.wiredTiger.concurrentTransactions.monitor.totalTickets": int32(128),
		"serverStatus.wiredTiger.concurrentTransactions.monitor.queueLength":  int32(0),
	}

	for key, wantVal := range want {
		gotVal, ok := got[key]
		if !ok {
			t.Fatalf("missing key %q; got %#v", key, got)
		}
		if gotVal != wantVal {
			t.Fatalf("key %q: got %v (%T), want %v (%T)", key, gotVal, gotVal, wantVal, wantVal)
		}
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected extra keys: %#v", got)
	}
}

func TestNormalizeMetricsDocument_flatQueueExecutionKeys(t *testing.T) {
	doc := birch.NewDocument(
		birch.EC.Int32("serverStatus.queues.execution.read.available", 100),
		birch.EC.Int32("serverStatus.queues.execution.write.available", 95),
		birch.EC.Int32("serverStatus.queues.execution.read.totalTickets", 128),
	)

	include := map[string]struct{}{
		"serverStatus.queues.execution.read.available":  {},
		"serverStatus.queues.execution.write.available": {},
		"serverStatus.queues.execution.read.totalTickets": {},
	}

	got := normalizeMetricsDocument(doc, include)
	if got["serverStatus.queues.execution.read.available"] != int32(100) {
		t.Fatalf("read.available = %#v", got)
	}
}

func TestNormalizeDocument_preservesNestedMetadata(t *testing.T) {
	doc := birch.NewDocument(
		birch.EC.SubDocument("doc", birch.NewDocument(
			birch.EC.SubDocument("hostInfo", birch.NewDocument(
				birch.EC.SubDocument("system", birch.NewDocument(
					birch.EC.String("hostname", "mongod1.example:27017"),
				)),
			)),
			birch.EC.SubDocument("buildInfo", birch.NewDocument(
				birch.EC.String("version", "8.0.17-6"),
			)),
		)),
	)

	got := normalizeDocument(doc, map[string]struct{}{})
	hostname := getNestedString(got, "doc.hostInfo.system.hostname")
	if hostname != "mongod1.example:27017" {
		t.Fatalf("hostname = %q, want mongod1.example:27017", hostname)
	}
	version := getNestedString(got, "doc.buildInfo.version")
	if version != "8.0.17-6" {
		t.Fatalf("version = %q, want 8.0.17-6", version)
	}
}

func TestNormalizeMetricsDocument_unwrapsTopLevelShardRole(t *testing.T) {
	doc := birch.NewDocument(
		birch.EC.SubDocument("shard", birch.NewDocument(
			birch.EC.SubDocument("replSetGetStatus", birch.NewDocument(
				birch.EC.ArrayFromElements("members",
					birch.VC.DocumentFromElements(
						birch.EC.Int32("state", 1),
						birch.EC.Int32("health", 1),
						birch.EC.Int32("pingMs", 2),
					),
				),
			)),
			birch.EC.SubDocument("local", birch.NewDocument(
				birch.EC.SubDocument("oplog", birch.NewDocument(
					birch.EC.SubDocument("rs", birch.NewDocument(
						birch.EC.SubDocument("stats", birch.NewDocument(
							birch.EC.SubDocument("storageStats", birch.NewDocument(
								birch.EC.Int64("storageSize", 1024),
							)),
						)),
					)),
				)),
			)),
		)),
	)

	include := map[string]struct{}{
		"replSetGetStatus.members.0.state":                        {},
		"replSetGetStatus.members.0.health":                       {},
		"replSetGetStatus.members.0.pingMs":                       {},
		"local.oplog.rs.stats.storageStats.storageSize":           {},
		"shard.replSetGetStatus.members.0.state":                   {}, // must not be required
	}

	got := normalizeMetricsDocument(doc, include)

	if got["replSetGetStatus.members.0.state"] != int32(1) {
		t.Fatalf("state = %#v", got["replSetGetStatus.members.0.state"])
	}
	if got["replSetGetStatus.members.0.health"] != int32(1) {
		t.Fatalf("health = %#v", got["replSetGetStatus.members.0.health"])
	}
	if got["replSetGetStatus.members.0.pingMs"] != int32(2) {
		t.Fatalf("pingMs = %#v", got["replSetGetStatus.members.0.pingMs"])
	}
	if got["local.oplog.rs.stats.storageStats.storageSize"] != int64(1024) {
		t.Fatalf("storageSize = %#v", got["local.oplog.rs.stats.storageStats.storageSize"])
	}
	if _, ok := got["shard.replSetGetStatus.members.0.state"]; ok {
		t.Fatalf("unexpected shard-prefixed key in output: %#v", got)
	}
}

func TestNormalizeMetricsDocument_roleCollisionKeepsTopLevelFirst(t *testing.T) {
	doc := birch.NewDocument(
		birch.EC.SubDocument("serverStatus", birch.NewDocument(
			birch.EC.Int32("uptime", 100),
		)),
		birch.EC.SubDocument("shard", birch.NewDocument(
			birch.EC.SubDocument("serverStatus", birch.NewDocument(
				birch.EC.Int32("uptime", 999),
			)),
		)),
	)

	include := map[string]struct{}{
		"serverStatus.uptime": {},
	}

	got := normalizeMetricsDocument(doc, include)
	if got["serverStatus.uptime"] != int32(100) {
		t.Fatalf("uptime = %#v, want top-level 100 (first-write-wins)", got["serverStatus.uptime"])
	}
}

func TestNormalizeMetricsDocument_nestedShardFieldNotUnwrapped(t *testing.T) {
	// Salvaguarda: "shard" nested under a non-empty prefix must remain in the path.
	doc := birch.NewDocument(
		birch.EC.SubDocument("serverStatus", birch.NewDocument(
			birch.EC.SubDocument("shard", birch.NewDocument(
				birch.EC.Int32("something", 7),
			)),
		)),
	)

	include := map[string]struct{}{
		"serverStatus.shard.something": {},
	}

	got := normalizeMetricsDocument(doc, include)
	if got["serverStatus.shard.something"] != int32(7) {
		t.Fatalf("got %#v", got)
	}
	if _, ok := got["something"]; ok {
		t.Fatalf("nested shard was incorrectly unwrapped: %#v", got)
	}
}

func TestNormalizeMetricsDocument_stripsFlatShardDottedKeys(t *testing.T) {
	// github.com/mongodb/ftdc often emits already-flattened dotted keys at the root.
	doc := birch.NewDocument(
		birch.EC.Int32("shard.replSetGetStatus.members.0.state", 1),
		birch.EC.Int64("shard.local.oplog.rs.stats.storageStats.storageSize", 2048),
		birch.EC.Int32("serverStatus.uptime", 10),
	)

	include := map[string]struct{}{
		"replSetGetStatus.members.0.state":              {},
		"local.oplog.rs.stats.storageStats.storageSize": {},
		"serverStatus.uptime":                           {},
	}

	got := normalizeMetricsDocument(doc, include)
	if got["replSetGetStatus.members.0.state"] != int32(1) {
		t.Fatalf("state = %#v", got["replSetGetStatus.members.0.state"])
	}
	if got["local.oplog.rs.stats.storageStats.storageSize"] != int64(2048) {
		t.Fatalf("storageSize = %#v", got["local.oplog.rs.stats.storageStats.storageSize"])
	}
	if _, ok := got["shard.replSetGetStatus.members.0.state"]; ok {
		t.Fatalf("shard-prefixed key remained: %#v", got)
	}
}

func TestNormalizeMetricsDocument_flatPre80LayoutUnchanged(t *testing.T) {
	// MongoDB 5/6/7 style: collectors at the top level, no role wrappers.
	doc := birch.NewDocument(
		birch.EC.SubDocument("replSetGetStatus", birch.NewDocument(
			birch.EC.ArrayFromElements("members",
				birch.VC.DocumentFromElements(
					birch.EC.Int32("state", 2),
				),
			),
		)),
		birch.EC.SubDocument("serverStatus", birch.NewDocument(
			birch.EC.Int32("uptime", 42),
		)),
	)

	include := map[string]struct{}{
		"replSetGetStatus.members.0.state": {},
		"serverStatus.uptime":             {},
	}

	got := normalizeMetricsDocument(doc, include)
	if got["replSetGetStatus.members.0.state"] != int32(2) {
		t.Fatalf("state = %#v", got["replSetGetStatus.members.0.state"])
	}
	if got["serverStatus.uptime"] != int32(42) {
		t.Fatalf("uptime = %#v", got["serverStatus.uptime"])
	}
	if len(got) != 2 {
		t.Fatalf("unexpected keys: %#v", got)
	}
}
