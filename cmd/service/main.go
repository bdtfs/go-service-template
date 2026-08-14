package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bdtfs/go-service-template/internal/di"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := di.Run(ctx, "config.yaml"); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}
