/**
MIT License

Copyright (c) 2023 Remko Tronçon

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
**/

package cslog

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const (
	LevelCritical = slog.Level(12)
)

// CloudLoggingOptions return options for writing structured Google logging entries
// https://github.com/remko/cloudrun-slog
func CloudLoggingOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			} else if a.Key == slog.SourceKey {
				a.Key = "logging.googleapis.com/sourceLocation"
			} else if a.Key == slog.LevelKey {
				a.Key = "severity"
				level := a.Value.Any().(slog.Level)
				if level == LevelCritical {
					a.Value = slog.StringValue("CRITICAL")
				}
			}
			return a
		}}
}

// Handler that outputs JSON understood by the structured log agent.
// See https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields
type CloudLoggingHandler struct{ handler slog.Handler }

// NewCloudLoggingHandler return a new Handler that outputs JSON understood by the structured log agent.
func NewCloudLoggingHandler() *CloudLoggingHandler {
	return &CloudLoggingHandler{handler: slog.NewJSONHandler(os.Stderr, CloudLoggingOptions())}
}

// Enabled implements slog.Handler.Enabled
func (h *CloudLoggingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.Handle
func (h *CloudLoggingHandler) Handle(ctx context.Context, rec slog.Record) error {
	trace := traceFromContext(ctx)
	if trace != "" {
		rec = rec.Clone()
		// Add trace ID	to the record so it is correlated with the Cloud Run request log
		// See https://cloud.google.com/trace/docs/trace-log-integration
		rec.Add("logging.googleapis.com/trace", slog.StringValue(trace))
	}
	return h.handler.Handle(ctx, rec)
}

// WithAttrs implements slog.Handler.WithAttrs
func (h *CloudLoggingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CloudLoggingHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup implements slog.Handler.WithGroup
func (h *CloudLoggingHandler) WithGroup(name string) slog.Handler {
	return &CloudLoggingHandler{handler: h.handler.WithGroup(name)}
}

// unique private key
var traceKey = struct{}{}

// ProjectID is required for the trace entry.
var ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

// WithCloudTraceContext is HTTP Middleware that adds the Cloud Trace ID to the context
// This is used to correlate the structured logs with the Cloud Run request log.
func WithCloudTraceContext(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var trace string
		traceHeader := r.Header.Get("X-Cloud-Trace-Context")
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", ProjectID, traceParts[0])
		}
		if trace == "" {
			h.ServeHTTP(w, r)
			return
		}
		// Pass in the default slog logger
		ctxlog := WithLogger(r.Context(), slog.Default())
		h.ServeHTTP(w, r.WithContext(context.WithValue(ctxlog, traceKey, trace)))
	})
}

func traceFromContext(ctx context.Context) string {
	trace := ctx.Value(traceKey)
	if trace == nil {
		return ""
	}
	return trace.(string)
}
