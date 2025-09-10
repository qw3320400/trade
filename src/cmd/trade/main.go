package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"trade/src/common"
	"trade/src/strategy"
)

func main() {
	common.InitLogger(false)
	common.Logger.Sugar().Info("Starting trade application...")

	err := strategy.NewPriceGap().Run(context.Background())
	if err != nil {
		common.Logger.Sugar().Fatalf("Failed to start PriceGap strategy: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		return
	}
}
