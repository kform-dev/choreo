package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands"
	"github.com/kform-dev/choreo/cmd/choreoctl/globals"
)

func main() {
	os.Exit(runMain())
}

// runMain does the initial setup to setup logging
func runMain() int {
	// init logging
	l := log.NewLogger(&log.HandlerOptions{Name: "choreoctl-logger", AddSource: false, MinLevel: globals.LogLevel})
	slog.SetDefault(l)

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx = log.IntoContext(ctx, l)

	// init cmd context
	cmd, f := commands.GetMain(ctx)
	defer f.Close()

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s \n", err.Error())
		cancel()
		return 1
	}
	return 0
}
