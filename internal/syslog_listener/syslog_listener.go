package syslog_listener

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"dnsmasq_exporter/internal/dnsmasq"
	"github.com/stratg5/go-syslog/format/v3"
	"github.com/stratg5/go-syslog/syslog/v3"
)

type Config struct {
	Port int
}

func Init(ctx context.Context, config Config, inputCh chan dnsmasq.Log) error {
	rawCh := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(rawCh)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC6587)
	server.SetHandler(handler)
	address := fmt.Sprintf("0.0.0.0:%v", config.Port)
	if err := server.ListenTCP(address); err != nil {
		return fmt.Errorf("SyslogListener: listen port failed, %v", err.Error())
	}
	if err := server.Boot(ctx); err != nil {
		return fmt.Errorf(fmt.Sprintf("SyslogListener: server boot failed, %v", err.Error()))
	}
	slog.InfoContext(ctx, fmt.Sprintf("SyslogListener: start, listen on tcp %v", address))

	go server.Wait()

	go func() {
		for {
			select {
			case logParts := <-rawCh:
				sessionCtx := context.Background()
				if appName, ok := logParts["app_name"].(string); ok {
					if appName != "dnsmasq" {
						break
					}
				}
				log := dnsmasq.Log{
					Context:   sessionCtx,
					Timestamp: extractTimestamp(logParts),
					Message:   extractMessage(logParts),
				}
				if log.Message == "" {
					break
				}
				log.Message = strings.TrimSuffix(log.Message, "\n")

				inputCh <- log
			case <-ctx.Done():
				if err := server.Kill(); err != nil {
					slog.Error(fmt.Sprintf("SyslogListener: kill server failed, %v", err.Error()))
				}
			}
		}
	}()
	return nil
}

func extractTimestamp(logParts format.LogParts) time.Time {
	if value, ok := logParts["timestamp"].(time.Time); ok {
		return value
	}
	return time.Now()
}

func extractMessage(logParts format.LogParts) string {
	if value, ok := logParts["message"].(string); ok {
		return value
	}
	return ""
}
