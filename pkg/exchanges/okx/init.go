package okx

import (
	"fmt"

	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("okx", func(cfg *config.Config) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.OKX.ResolveAccount("")
		if !ok {
			return nil, fmt.Errorf("okx: no accounts configured")
		}
		return NewOKXExchange(acc, cfg.Exchanges.OKX.Testnet)
	})
	exchanges.RegisterAccountFactory("okx", func(cfg *config.Config, accountName string) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.OKX.ResolveAccount(accountName)
		if !ok {
			return nil, fmt.Errorf("okx: account %q not found", accountName)
		}
		return NewOKXExchange(acc, cfg.Exchanges.OKX.Testnet)
	})
}
