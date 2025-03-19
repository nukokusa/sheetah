package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nukokusa/sheetah"
	"golang.org/x/sys/unix"
)

var version = "current"

func main() {
	sheetah.Version = version
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, unix.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	c, err := sheetah.New(ctx)
	if err != nil {
		return err
	}
	return c.Run(ctx)
}
