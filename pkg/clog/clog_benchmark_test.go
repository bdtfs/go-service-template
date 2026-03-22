package clog_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/bdtfs/go-service-template/pkg/clog"
)

func BenchmarkCustomLogger(b *testing.B) {
	var buf bytes.Buffer

	logger := clog.NewCLog(slog.LevelDebug, &buf, false)

	ctx := logger.AddKeysValuesToCtx(context.Background(), map[string]any{
		"userID":    12345,
		"userName":  "testuser",
		"timestamp": time.Now(),
		"data":      []int{1, 2, 3},
	})

	b.ResetTimer()
	for b.Loop() {
		logger.InfoCtx(ctx, "Some test message")
	}
}
