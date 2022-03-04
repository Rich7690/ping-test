package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rich7690/ping-test/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	internal.StartServer(ctx, sigs, ":2112")
	<-sigs
	cancel()
	log.Info().Msg("Exiting")
}
