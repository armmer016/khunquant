package binanceth

import (
	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("binanceth", func(cfg *config.Config) (exchanges.Exchange, error) {
		return NewBinanceTHExchange(cfg)
	})
}
