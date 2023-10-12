package cslog

import (
	"context"
	"testing"

	"log/slog"
)

func TestWithLogger(t *testing.T) {
	mylog := slog.New(slog.Default().Handler()).With(slog.String("mykey", "myvalue"))
	ctx := WithLogger(context.Background(), mylog)
	hislog := Ctx(ctx)
	hislog.Info("Hello, world!")
}
