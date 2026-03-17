package binance

import (
	"fmt"

	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("binance", func(cfg *config.Config) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.Binance.ResolveAccount("")
		if !ok {
			return nil, fmt.Errorf("binance: no accounts configured")
		}
		return NewBinanceExchange(acc, cfg.Exchanges.Binance.Testnet)
	})
	exchanges.RegisterAccountFactory("binance", func(cfg *config.Config, accountName string) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.Binance.ResolveAccount(accountName)
		if !ok {
			return nil, fmt.Errorf("binance: account %q not found", accountName)
		}
		return NewBinanceExchange(acc, cfg.Exchanges.Binance.Testnet)
	})
}
