package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func regSignalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(done)
		<-done
		log.Info().Msg("stop signal received")
		cancel()
	}()

	return ctx
}

func runServer(ctx context.Context) error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Info().Msg("server started")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error().Err(err).Msg("error accepting")
				return
			}
			go handleConnection(conn)
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("server was closed")
		break

	}
	return nil

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintf(conn, "welcome!\n")
	time.Sleep(time.Second)
	fmt.Fprintf(conn, "goodbye!\n")
}

func main() {
	ctx := regSignalHandler(context.Background())

	if err := runServer(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server stopped on error")
	}
}
