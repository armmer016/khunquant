package bitkub

import (
	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("bitkub", func(cfg *config.Config) (exchanges.Exchange, error) {
		return NewBitkubExchange(cfg)
	})
}
