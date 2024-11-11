package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/pashest/word-of-wisdom/config"
	"github.com/pashest/word-of-wisdom/internal/cache"
	"github.com/pashest/word-of-wisdom/internal/pkg/pow/equihash"
	"github.com/pashest/word-of-wisdom/internal/pkg/quotes"
	"github.com/pashest/word-of-wisdom/internal/server"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	equihashSetting, err := equihash.NewSetting(cfg)
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	quoteService := quotes.New(cfg)
	cache := cache.NewCache(time.Second * 10)

	srv := server.NewServer(cfg, cache, quoteService, equihashSetting)

	go func() {
		err := srv.Run(ctx)
		if err != nil {
			log.Fatal().Err(err)
		}
	}()

	<-ctx.Done()
	srv.Stop()
}
