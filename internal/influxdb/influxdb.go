package influxdb

import (
	"context"
	"fmt"
	"log/slog"

	"dnsmasq_exporter/internal/dnsmasq"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Config struct {
	Url    string
	Token  string
	Org    string
	Bucket string
}

var (
	client influxdb2.Client
)

func Init(ctx context.Context, config Config, outputCh chan dnsmasq.Point) error {
	client = influxdb2.NewClient(config.Url, config.Token)
	writeAPI := client.WriteAPIBlocking(config.Org, config.Bucket)

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := writeAPI.Flush(ctx)
				if err != nil {
					slog.Error(fmt.Sprintf("InfluxDB: flush writeAPI failed, %v", err.Error()))
				}
				client.Close()
				return
			case point, ok := <-outputCh:
				if !ok {
					return
				}
				p := influxdb2.NewPoint(
					"aries.dnsmasq",
					map[string]string{
						"action": point.Action,
						"type":   point.Type,
						"domain": point.Domain,
						"from":   point.From,
					},
					map[string]any{
						"count": 1,
					},
					point.Timestamp,
				)
				if err := writeAPI.WritePoint(point.Context, p); err != nil {
					slog.Error("InfluxDB write point failed", "err", err)
				}
			}
		}
	}()
	return nil
}
