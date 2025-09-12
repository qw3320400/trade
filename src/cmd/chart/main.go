package main

import (
	"context"
	"trade/src/chart"
	"trade/src/common"
)

func main() {
	components := []common.Component{
		chart.NewPriceGapChart(),
	}
	for _, component := range components {
		if err := component.Run(context.Background()); err != nil {
			common.Logger.Sugar().Fatalf("Failed to run component: %v", err)
		}
	}
}
