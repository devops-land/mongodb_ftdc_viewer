package influx

import (
	"context"
	"fmt"
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

var grafanaDateFormat = "2006-01-02T15:04:05.000Z"

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

func (i *Client) FetchEarliestTimestamp() (error, string) {
	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: -1y)  // Adjust this range if necessary
  |> filter(fn: (r) => r._measurement == "%s")
  |> filter(fn: (r) => r._field == "start")
  |> group()
  |> first()
  |> yield(name: "first_value")`, i.config.Bucket, i.config.Measurement)

	result, err := i.client.QueryAPI(i.config.Org).Query(context.Background(), query)
	if err != nil {
		return err, ""
	}

	for result.Next() {
		if result.Err() != nil {
			return result.Err(), ""
		}

		t, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", result.Record().Time().String())

		return nil, t.Format(grafanaDateFormat)
	}

	return fmt.Errorf("no earliest timestamp"), ""
}

func (i *Client) FetchLatestTimestamp() (error, string) {
	query := fmt.Sprintf(`
from(bucket: "%s")
  |> range(start: -1y)  // Adjust this range if necessary
  |> filter(fn: (r) => r._measurement == "%s")
  |> filter(fn: (r) => r._field == "start")
  |> group()
  |> last()
  |> yield(name: "last_value")`, i.config.Bucket, i.config.Measurement)

	result, err := i.client.QueryAPI(i.config.Org).Query(context.Background(), query)
	if err != nil {
		return err, ""
	}

	for result.Next() {
		if result.Err() != nil {
			return result.Err(), ""
		}

		t, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", result.Record().Time().String())

		return nil, t.Format(grafanaDateFormat)
	}

	return fmt.Errorf("no latest timestamp"), ""
}
