package cslog

import (
	"context"
	"errors"
	"testing"

	"log/slog"
)

func TestWithLogger(t *testing.T) {
	mylog := slog.New(slog.Default().Handler()).With(slog.String("mykey", "myvalue"))
	ctx := WithLogger(context.Background(), mylog)
	hislog := Ctx(ctx)
	hislog.Info("Hello, world!")
	err := errors.New("bummer")
	hislog.Error("Panic", "err", err, slog.Any("err", err))
}
