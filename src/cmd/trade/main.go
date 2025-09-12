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

	components := []common.Component{
		strategy.NewPriceGap(),
	}
	for _, component := range components {
		if err := component.Run(context.Background()); err != nil {
			common.Logger.Sugar().Fatalf("Failed to run component: %v", err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
		return
	}
}
