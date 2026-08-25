package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"godruid/internal/app"
	"godruid/internal/logx"
)

func main() {
	settings := app.LoadSettings()
	log := logx.Init(settings.LogLevel, os.Stdout)
	instance, err := app.New(settings, log)
	if err != nil {
		log.Error("init failed", "err", err)
		os.Exit(1)
	}
	if err := instance.Start(); err != nil {
		log.Error("start failed", "err", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		log.Error("shutdown", "err", err)
		os.Exit(1)
	}
}
