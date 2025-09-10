package strategy

import (
	"context"
	"strconv"
	"time"
	"trade/src/common"
	"trade/src/exchange/binance"
	"trade/src/exchange/okx"

	"github.com/shopspring/decimal"
)

type Pair struct {
	BinancePrice *ExchangePrice
	OKXPrice     *ExchangePrice
}

type ExchangePrice struct {
	Symbol    string
	MarkPrice decimal.Decimal
	Time      int64
}

type PriceGap struct {
	bCli  *binance.Client
	oCli  *okx.Client
	pairs []*Pair
}

func NewPriceGap() *PriceGap {
	return &PriceGap{
		bCli: binance.NewClient(),
		oCli: okx.NewClient(),
		pairs: []*Pair{
			{
				BinancePrice: &ExchangePrice{Symbol: "BTCUSDT"},
				OKXPrice:     &ExchangePrice{Symbol: "BTC-USDT-SWAP"},
			},
			{
				BinancePrice: &ExchangePrice{Symbol: "ETHUSDT"},
				OKXPrice:     &ExchangePrice{Symbol: "ETH-USDT-SWAP"},
			},
			{
				BinancePrice: &ExchangePrice{Symbol: "SOLUSDT"},
				OKXPrice:     &ExchangePrice{Symbol: "SOL-USDT-SWAP"},
			},
			{
				BinancePrice: &ExchangePrice{Symbol: "DOGEUSDT"},
				OKXPrice:     &ExchangePrice{Symbol: "DOGE-USDT-SWAP"},
			},
			{
				BinancePrice: &ExchangePrice{Symbol: "XRPUSDT"},
				OKXPrice:     &ExchangePrice{Symbol: "XRP-USDT-SWAP"},
			},
			{
				BinancePrice: &ExchangePrice{Symbol: "WLFIUSDT"},
				OKXPrice:     &ExchangePrice{Symbol: "WLFI-USDT-SWAP"},
			},
		},
	}
}

func (p *PriceGap) Run(ctx context.Context) error {
	err := p.bCli.InitFuturesStreamWebSocketConnection(ctx)
	if err != nil {
		return err
	}
	err = p.oCli.InitPublicWebSocketConnection(ctx)
	if err != nil {
		return err
	}
	for _, pair := range p.pairs {
		go p.RunPair(ctx, pair)
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (p *PriceGap) RunPair(ctx context.Context, pair *Pair) {
	defer common.HandlePanic()
	common.Logger.Sugar().Infof("PriceGap RunPair %s %s", pair.BinancePrice.Symbol, pair.OKXPrice.Symbol)

	bPrice := binance.NewFuturesStreamWebSocketMarketPrice(pair.BinancePrice.Symbol)
	err := p.bCli.Subscribe(bPrice.Subscribe(), func(stream *binance.FuturesStreamWebSocketStream) {
		_, err := bPrice.Stream(stream)
		if err != nil {
			common.Logger.Sugar().Warnf("PriceGap RunPair %s Binance Stream error: %v", pair.BinancePrice.Symbol, err)
			return
		}
		pair.BinancePrice.MarkPrice = bPrice.MarkPrice
		pair.BinancePrice.Time = bPrice.EventTime
	})
	if err != nil {
		common.Logger.Sugar().Errorf("PriceGap RunPair %s Subscribe Binance error: %v", pair.BinancePrice.Symbol, err)
		return
	}

	oPrice := okx.NewPublicWebSocketMarkPrices(pair.OKXPrice.Symbol)
	err = p.oCli.Subscribe(oPrice.Subscribe(), func(stream *okx.WebSocketStream) {
		_, err := oPrice.Stream(stream)
		if err != nil {
			common.Logger.Sugar().Warnf("PriceGap RunPair %s OKX Stream error: %v", pair.OKXPrice.Symbol, err)
			return
		}
		var maxTime *okx.PublicWebSocketMarkPrice
		for _, p := range *oPrice {
			if maxTime == nil || p.Timestamp > maxTime.Timestamp {
				maxTime = p
			}
		}
		if maxTime != nil {
			pair.OKXPrice.MarkPrice = maxTime.MarkPrice
			pair.OKXPrice.Time, _ = strconv.ParseInt(maxTime.Timestamp, 10, 64)
		}
	})
	if err != nil {
		common.Logger.Sugar().Errorf("PriceGap RunPair %s Subscribe OKX error: %v", pair.OKXPrice.Symbol, err)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.checkPriceGap(pair)
		case <-ctx.Done():
			return
		}
	}
}

func (p *PriceGap) checkPriceGap(pair *Pair) {
	if pair.BinancePrice.MarkPrice.IsZero() || pair.OKXPrice.MarkPrice.IsZero() {
		return
	}
	if pair.BinancePrice.Time == 0 || pair.OKXPrice.Time == 0 {
		return
	}
	gap := pair.BinancePrice.MarkPrice.Sub(pair.OKXPrice.MarkPrice)
	avg := pair.BinancePrice.MarkPrice.Add(pair.OKXPrice.MarkPrice).Div(decimal.NewFromInt(2))
	ratio := gap.Div(avg).Mul(decimal.NewFromInt(100))
	ratioAbs := ratio.Abs()
	var logFunc func(msg string, args ...interface{})
	if ratioAbs.GreaterThanOrEqual(decimal.NewFromFloat(0.03)) {
		logFunc = common.Logger.Sugar().Infof
	} else if ratioAbs.GreaterThanOrEqual(decimal.NewFromFloat(0.1)) {
		logFunc = common.Logger.Sugar().Warnf
	} else if ratioAbs.GreaterThanOrEqual(decimal.NewFromFloat(0.3)) {
		logFunc = common.Logger.Sugar().Errorf
	}
	if logFunc != nil {
		now := time.Now().UnixMilli()
		logFunc("PriceGap %s(%s/%dms) %s(%s/%dms) gap: %s ratio: %s%%",
			pair.BinancePrice.Symbol, pair.BinancePrice.MarkPrice.String(), now-pair.BinancePrice.Time,
			pair.OKXPrice.Symbol, pair.OKXPrice.MarkPrice.String(), now-pair.OKXPrice.Time,
			gap.String(), ratio.StringFixed(4))
	}
}
