package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/27149chen/dex-example/pkg/extauth/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		stop()
	}()

	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}
