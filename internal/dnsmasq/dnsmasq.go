package dnsmasq

import (
	"context"
	"log/slog"
	"strings"
)

func Init(ctx context.Context, inputCh chan Log, outputCh chan Point) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-inputCh:
				if !ok {
					return
				}
				point, ok := parseMessage(ctx, message)
				if ok {
					outputCh <- point
				}
			}
		}
	}()
}

func parseMessage(_ context.Context, log Log) (Point, bool) {
	values := strings.Split(log.Message, " ")
	if len(values) != 6 {
		slog.WarnContext(log.Context, "Received unknown message: "+log.Message)
		return Point{}, false
	}
	//id := values[0]
	src := values[1]
	if index := strings.Index(src, "/"); index != -1 {
		src = src[:index]
	}
	action := values[2]
	domainName := values[3]
	from := values[5]

	if strings.HasPrefix(action, "query") {
		nsType := "-"
		if index := strings.Index(action, "["); index != -1 {
			nsType = action[index+1 : len(action)-1]
		}
		point := Point{
			Context:   log.Context,
			Timestamp: log.Timestamp,
			Action:    "query",
			Type:      nsType,
			Domain:    domainName,
			From:      from,
		}
		return point, true
	}
	return Point{}, false
}
