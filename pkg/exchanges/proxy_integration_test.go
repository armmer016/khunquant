package exchanges_test

// Proxy integration test — verifies that exchange HTTP clients correctly route
// traffic through a configured proxy. Run with:
//
//	ssh -D 1080 -N ubuntu@57.129.141.80 &
//	go test ./pkg/exchanges/ -run TestProxyRouting -v -tags proxy_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cryptoquantumwave/khunquant/pkg/config"
	"github.com/cryptoquantumwave/khunquant/pkg/exchanges"
	_ "github.com/cryptoquantumwave/khunquant/pkg/exchanges/binance"
	_ "github.com/cryptoquantumwave/khunquant/pkg/exchanges/bitkub"
	_ "github.com/cryptoquantumwave/khunquant/pkg/exchanges/binanceth"
	_ "github.com/cryptoquantumwave/khunquant/pkg/exchanges/okx"
)

func proxyOrSkip(t *testing.T) string {
	t.Helper()
	proxy := os.Getenv("EXCHANGE_PROXY")
	if proxy == "" {
		t.Skip("EXCHANGE_PROXY not set — set to e.g. socks5://127.0.0.1:1080 or http://127.0.0.1:3128")
	}
	return proxy
}

func publicCfg(proxy string) *config.Config {
	cfg := &config.Config{}
	acc := config.ExchangeAccount{Proxy: proxy}
	cfg.Exchanges.Binance.Enabled = true
	cfg.Exchanges.Binance.Accounts = []config.ExchangeAccount{acc}
	cfg.Exchanges.Bitkub.Enabled = true
	cfg.Exchanges.Bitkub.Accounts = []config.ExchangeAccount{acc}
	cfg.Exchanges.BinanceTH.Enabled = true
	cfg.Exchanges.BinanceTH.Accounts = []config.ExchangeAccount{acc}
	cfg.Exchanges.OKX.Enabled = true
	okxAcc := config.OKXExchangeAccount{ExchangeAccount: acc}
	cfg.Exchanges.OKX.Accounts = []config.OKXExchangeAccount{okxAcc}
	return cfg
}

func fetchPrice(t *testing.T, ex exchanges.Exchange, ctx context.Context, asset, quote string) float64 {
	t.Helper()
	pe, ok := ex.(exchanges.PricedExchange)
	if !ok {
		t.Fatalf("%s does not implement PricedExchange", ex.Name())
	}
	price, err := pe.FetchPrice(ctx, asset, quote)
	if err != nil {
		t.Fatalf("fetch price: %v", err)
	}
	return price
}

func TestProxyRouting_Bitkub(t *testing.T) {
	proxy := proxyOrSkip(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ex, err := exchanges.CreateExchange("bitkub", publicCfg(proxy))
	if err != nil {
		t.Fatalf("create bitkub: %v", err)
	}
	price := fetchPrice(t, ex, ctx, "BTC", "THB")
	if price <= 0 {
		t.Fatalf("unexpected price: %v", price)
	}
	t.Logf("OK  bitkub  BTC/THB = %.2f  proxy=%s", price, proxy)
}

func TestProxyRouting_BinanceTH(t *testing.T) {
	proxy := proxyOrSkip(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ex, err := exchanges.CreateExchange("binanceth", publicCfg(proxy))
	if err != nil {
		t.Fatalf("create binanceth: %v", err)
	}
	price := fetchPrice(t, ex, ctx, "BTC", "THB")
	if price <= 0 {
		t.Fatalf("unexpected price: %v", price)
	}
	t.Logf("OK  binanceth  BTC/THB = %.2f  proxy=%s", price, proxy)
}

func TestProxyRouting_Binance(t *testing.T) {
	proxy := proxyOrSkip(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ex, err := exchanges.CreateExchange("binance", publicCfg(proxy))
	if err != nil {
		t.Fatalf("create binance: %v", err)
	}
	price := fetchPrice(t, ex, ctx, "BTC", "USDT")
	if price <= 0 {
		t.Fatalf("unexpected price: %v", price)
	}
	t.Logf("OK  binance  BTC/USDT = %.2f  proxy=%s", price, proxy)
}

func TestProxyRouting_OKX(t *testing.T) {
	proxy := proxyOrSkip(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ex, err := exchanges.CreateExchange("okx", publicCfg(proxy))
	if err != nil {
		t.Fatalf("create okx: %v", err)
	}
	price := fetchPrice(t, ex, ctx, "BTC", "USDT")
	if price <= 0 {
		t.Fatalf("unexpected price: %v", price)
	}
	t.Logf("OK  okx  BTC/USDT = %.2f  proxy=%s", price, proxy)
}
