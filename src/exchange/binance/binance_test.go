package binance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBinance(t *testing.T) {
	t.Run("FuturesAPI", func(t *testing.T) {
		testFuturesAPIPriceTicker(t)
	})
	t.Run("FuturesStream", func(t *testing.T) {
		testFuturesMarketPriceStream(t)
	})
}

func testFuturesAPIPriceTicker(t *testing.T) {
	cli := NewClient()
	defer cli.Clean()
	err := cli.InitFuturesAPIWebSocketConnection(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ticker := NewFuturesAPIWebSocketPriceTicker("BTCUSDT")
	resp, err := cli.Call(ticker.Request())
	if err != nil {
		t.Fatal(err)
	}
	_, err = ticker.Response(resp)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "BTCUSDT", ticker.Symbol)
	require.NotZero(t, ticker.Price)
	require.NotZero(t, ticker.Time)
}

func testFuturesMarketPriceStream(t *testing.T) {
	cli := NewClient()
	defer cli.Clean()
	var count int
	err := cli.InitFuturesStreamWebSocketConnection(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	price := NewFuturesStreamWebSocketMarketPrice("BTCUSDT")
	err = cli.Subscribe(price.Subscribe(), func(stream *FuturesStreamWebSocketStream) {
		_, err := price.Stream(stream)
		if err != nil {
			t.Logf("FuturesMarketPriceStream Stream error: %v", err)
			return
		}
		require.Equal(t, "BTCUSDT", price.Symbol)
		require.NotZero(t, price.MarkPrice)
		require.NotZero(t, price.EventTime)
		count++
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	require.Greater(t, count, 0)
}
