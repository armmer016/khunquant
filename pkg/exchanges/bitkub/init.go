package bitkub

import (
	"fmt"

	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("bitkub", func(cfg *config.Config) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.Bitkub.ResolveAccount("")
		if !ok {
			return nil, fmt.Errorf("bitkub: no accounts configured")
		}
		return NewBitkubExchange(acc)
	})
	exchanges.RegisterAccountFactory("bitkub", func(cfg *config.Config, accountName string) (exchanges.Exchange, error) {
		acc, ok := cfg.Exchanges.Bitkub.ResolveAccount(accountName)
		if !ok {
			return nil, fmt.Errorf("bitkub: account %q not found", accountName)
		}
		return NewBitkubExchange(acc)
	})
}
