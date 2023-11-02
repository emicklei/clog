package cslog

import (
	"os"
	"testing"

	"log/slog"
)

func TestCloudLogging(t *testing.T) {
	h := slog.NewJSONHandler(os.Stderr, CloudLoggingOptions())
	slog.SetDefault(slog.New(h))
	slog.Info("Hello, world!")
}
