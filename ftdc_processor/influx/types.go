package influx

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"time"
)

type Client struct {
	ctx    context.Context
	client influxdb2.Client
	api    api.WriteAPIBlocking
	config Config
}

type Config struct {
	Org         string
	Bucket      string
	Url         string
	Token       string
	Measurement string
	UseGzip     bool
}

func (i *Client) Close() {
	i.client.Close()
}

func (i *Client) WritePoint(point ...*write.Point) error {
	return i.api.WritePoint(i.ctx, point...)
}

func (i *Client) NewPoint(
	tags map[string]string,
	doc map[string]interface{},
	ts time.Time) *Point {
	return influxdb2.NewPoint(i.config.Measurement, tags, doc, ts)
}

type Point = write.Point
