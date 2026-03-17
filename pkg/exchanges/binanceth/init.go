package binanceth

import (
	"fmt"

	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("binanceth", func(cfg *config.Config) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.BinanceTH.ResolveAccount("")
		if !ok {
			return nil, fmt.Errorf("binanceth: no accounts configured")
		}
		return NewBinanceTHExchange(acc)
	})
	exchanges.RegisterAccountFactory("binanceth", func(cfg *config.Config, accountName string) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.BinanceTH.ResolveAccount(accountName)
		if !ok {
			return nil, fmt.Errorf("binanceth: account %q not found", accountName)
		}
		return NewBinanceTHExchange(acc)
	})
}
