package main

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// logLevel converts vague log level name into typed level.
func logLevel(s string) log.Level {
	switch s {
	case "1", "error":
		return log.ErrorLevel
	case "2", "warn":
		return log.WarnLevel
	case "3", "info":
		return log.InfoLevel
	case "4", "debug":
		return log.DebugLevel
	default:
		return log.FatalLevel
	}
}

// toBool is used to parse vague bool definition into typed bool.
func toBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "t", "yes":
		return true
	default:
		return false
	}
}

// duration parses duration from string with a provided default fallback.
func duration(s string, defaults time.Duration) time.Duration {
	dur, err := time.ParseDuration(s)
	if err != nil {
		dur = defaults
	}
	return dur
}

// checkStatsdPrefix ensures that the statsd prefix really
// have "." at end.
func checkStatsdPrefix(s string) string {
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}

func waitForService(ctx context.Context, conn *grpc.ClientConn) error {
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("Service wait timed out. Please run injective exchange service:\n\nmake install && injective-exchange")
		default:
			state := conn.GetState()

			if state != connectivity.Ready {
				time.Sleep(time.Second)
				continue
			}

			return nil
		}
	}
}
