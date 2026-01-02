package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"dnsmasq_exporter/internal/dnsmasq"
	"dnsmasq_exporter/internal/influxdb"
	"dnsmasq_exporter/internal/syslog_listener"
	"github.com/goccy/go-yaml"
)

type Config struct {
	SyslogNg syslog_listener.Config `yaml:"syslog_ng"`
	InfluxDB influxdb.Config        `yaml:"influxdb"`
}

func main() {
	config := loadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	inputCh := make(chan dnsmasq.Log, 1024)
	outputCh := make(chan dnsmasq.Point, 1024)

	dnsmasq.Init(ctx, inputCh, outputCh)

	if err := syslog_listener.Init(ctx, config.SyslogNg, inputCh); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := influxdb.Init(ctx, config.InfluxDB, outputCh); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	<-ctx.Done()
	slog.InfoContext(ctx, "Received shutdown signal, exiting gracefully...")
}

func loadConfig() *Config {
	configFilePath := flag.String("config", "./configs/config.yaml", "Config file path")
	flag.Parse()
	if configFilePath == nil || *configFilePath == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Config file not provide")
		os.Exit(2)
	}
	if _, err := os.Stat(*configFilePath); err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(os.Stderr, "Config file not exist")
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Stat config file failed, %v", err)
		}
		os.Exit(3)
	}

	configBytes, err := os.ReadFile(*configFilePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Read config file failed, %v", err)
		os.Exit(3)
	}
	var config Config
	if err = yaml.Unmarshal(configBytes, &config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unmarshal config file failed, %v", err)
		os.Exit(3)
	}
	return &config
}
