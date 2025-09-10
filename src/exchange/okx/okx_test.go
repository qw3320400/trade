package okx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOKX(t *testing.T) {
	t.Run("FuturesStream", func(t *testing.T) {
		testPublicWebSocketMarkPrices(t)
	})
}

func testPublicWebSocketMarkPrices(t *testing.T) {
	cli := NewClient()
	defer cli.Clean()
	var count int
	err := cli.InitPublicWebSocketConnection(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	price := NewPublicWebSocketMarkPrices("BTC-USDT-SWAP")
	err = cli.Subscribe(price.Subscribe(), func(stream *WebSocketStream) {
		_, err := price.Stream(stream)
		if err != nil {
			t.Logf("PublicWebSocketMarkPrices Stream error: %v %s", err, string(stream.Data))
			return
		}
		for _, p := range *price {
			require.Equal(t, "BTC-USDT-SWAP", p.InstID)
			require.NotZero(t, p.MarkPrice)
			require.NotEmpty(t, p.Timestamp)
			count++
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	require.Greater(t, count, 0)
}
