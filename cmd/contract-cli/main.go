package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/invocation"
)

func main() {
	if handled, err := invocation.RunInspectionHelper(os.Args[1:], os.Stdout); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	app := cli.New(cli.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Logger: logger,
	})

	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		logger.Error("command failed", "error", err.Error())
		os.Exit(1)
	}
}
