package influx

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"time"
)

func NewClient(ctx context.Context, cfg Config) Client {
	client := influxdb2.NewClientWithOptions(cfg.Url, cfg.Token, influxdb2.DefaultOptions().SetUseGZip(cfg.UseGzip).SetPrecision(time.Second).SetMaxRetries(5).SetMaxRetryInterval(10000))
	return Client{
		ctx:    ctx,
		client: client,
		api:    client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
	}
}
