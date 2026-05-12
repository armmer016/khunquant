package snapshot

import (
	"context"
	"testing"

	"github.com/cryptoquantumwave/khunquant/pkg/config"
	"github.com/cryptoquantumwave/khunquant/pkg/exchanges"
)

// --- sliceContains ---

func TestSliceContains_Found(t *testing.T) {
	if !sliceContains([]string{"a", "b", "c"}, "b") {
		t.Error("expected sliceContains to find 'b'")
	}
}

func TestSliceContains_NotFound(t *testing.T) {
	if sliceContains([]string{"a", "b", "c"}, "z") {
		t.Error("expected sliceContains to return false for missing element")
	}
}

func TestSliceContains_EmptySlice(t *testing.T) {
	if sliceContains(nil, "a") {
		t.Error("expected false for nil slice")
	}
}

// --- listExchangeAccounts ---

func TestListExchangeAccounts_Empty(t *testing.T) {
	cfg := &config.Config{}
	result := listExchangeAccounts(cfg)
	if len(result) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(result))
	}
}

func TestListExchangeAccounts_DisabledExchange(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Binance: config.BinanceExchangeConfig{
				Enabled:  false,
				Accounts: []config.ExchangeAccount{{Name: "main"}},
			},
		},
	}
	result := listExchangeAccounts(cfg)
	if len(result) != 0 {
		t.Errorf("expected 0 accounts for disabled exchange, got %d", len(result))
	}
}

func TestListExchangeAccounts_SingleEnabled(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Binance: config.BinanceExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "spot"}, {Name: "futures"}},
			},
		},
	}
	result := listExchangeAccounts(cfg)
	if len(result) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(result))
	}
	for _, ea := range result {
		if ea.exchange != "binance" {
			t.Errorf("expected exchange 'binance', got %q", ea.exchange)
		}
	}
	if result[0].account != "spot" || result[1].account != "futures" {
		t.Errorf("unexpected accounts: %+v", result)
	}
}

func TestListExchangeAccounts_MultipleExchanges(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Binance: config.BinanceExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "main"}},
			},
			Bitkub: config.BitkubExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "default"}},
			},
		},
	}
	result := listExchangeAccounts(cfg)
	if len(result) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(result))
	}
}

// --- effectiveQuote ---

// baseExchange is a minimal Exchange with no QuoteLister implementation.
type baseExchange struct{}

func (b *baseExchange) Name() string                                               { return "base" }
func (b *baseExchange) GetBalances(_ context.Context) ([]exchanges.Balance, error) { return nil, nil }

// quotedExchange implements QuoteLister on top of a basic exchange.
type quotedExchange struct {
	baseExchange
	quotes []string
}

func (q *quotedExchange) SupportedQuotes() []string { return q.quotes }

func TestEffectiveQuote_NoQuoteLister(t *testing.T) {
	ex := &baseExchange{}
	got := effectiveQuote(ex, "USDT")
	if got != "USDT" {
		t.Errorf("expected USDT, got %q", got)
	}
}

func TestEffectiveQuote_SupportedQuote(t *testing.T) {
	ex := &quotedExchange{quotes: []string{"THB", "USDT"}}
	got := effectiveQuote(ex, "USDT")
	if got != "USDT" {
		t.Errorf("expected USDT, got %q", got)
	}
}

func TestEffectiveQuote_UnsupportedFallsBackToFirst(t *testing.T) {
	ex := &quotedExchange{quotes: []string{"THB"}}
	got := effectiveQuote(ex, "USDT")
	// "USDT" is not in supported quotes, so fallback to first = "THB"
	if got != "THB" {
		t.Errorf("expected THB fallback, got %q", got)
	}
}

func TestEffectiveQuote_EmptyQuoteList(t *testing.T) {
	ex := &quotedExchange{quotes: []string{}}
	got := effectiveQuote(ex, "USDT")
	// No quotes available → return requested quote as-is
	if got != "USDT" {
		t.Errorf("expected USDT when quote list empty, got %q", got)
	}
}

func TestListExchangeAccounts_BinanceTH(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			BinanceTH: config.BinanceTHExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "th-main"}},
			},
		},
	}
	result := listExchangeAccounts(cfg)
	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}
	if result[0].exchange != "binanceth" || result[0].account != "th-main" {
		t.Errorf("unexpected account: %+v", result[0])
	}
}

func TestListExchangeAccounts_OKX(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			OKX: config.OKXExchangeConfig{
				Enabled:  true,
				Accounts: []config.OKXExchangeAccount{{ExchangeAccount: config.ExchangeAccount{Name: "okx-main"}}},
			},
		},
	}
	result := listExchangeAccounts(cfg)
	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}
	if result[0].exchange != "okx" || result[0].account != "okx-main" {
		t.Errorf("unexpected account: %+v", result[0])
	}
}

func TestListExchangeAccounts_Settrade(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Settrade: config.SettradeExchangeConfig{
				Enabled:  true,
				Accounts: []config.SettradeExchangeAccount{{ExchangeAccount: config.ExchangeAccount{Name: "st-main"}}},
			},
		},
	}
	result := listExchangeAccounts(cfg)
	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}
	if result[0].exchange != "settrade" || result[0].account != "st-main" {
		t.Errorf("unexpected account: %+v", result[0])
	}
}

func TestCollectFromExchanges_NoAccounts(t *testing.T) {
	cfg := &config.Config{}
	_, err := CollectFromExchanges(context.Background(), cfg, CollectOptions{})
	if err == nil {
		t.Error("CollectFromExchanges with no accounts should return error")
	}
}

func TestCollectFromExchanges_SourceFilterNoMatch(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Binance: config.BinanceExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "main"}},
			},
		},
	}
	_, err := CollectFromExchanges(context.Background(), cfg, CollectOptions{Source: "nonexistent-exchange"})
	if err == nil {
		t.Error("CollectFromExchanges with unmatched source filter should return error")
	}
}

func TestCollectFromExchanges_SourceAllMeansAllAccounts(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Binance: config.BinanceExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "main"}},
			},
		},
	}
	result, err := CollectFromExchanges(context.Background(), cfg, CollectOptions{Source: " all "})
	if err != nil {
		t.Fatalf("CollectFromExchanges source=all should not filter out accounts: %v", err)
	}
	if result == nil {
		t.Fatal("CollectFromExchanges source=all returned nil result")
	}
}

func TestCollectFromExchanges_SourceFilterAccountMismatch(t *testing.T) {
	cfg := &config.Config{
		Exchanges: config.ExchangesConfig{
			Binance: config.BinanceExchangeConfig{
				Enabled:  true,
				Accounts: []config.ExchangeAccount{{Name: "main"}},
			},
		},
	}
	_, err := CollectFromExchanges(context.Background(), cfg, CollectOptions{Source: "binance", Account: "nonexistent"})
	if err == nil {
		t.Error("CollectFromExchanges with unmatched account filter should return error")
	}
}
