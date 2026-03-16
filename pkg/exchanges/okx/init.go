package okx

import (
	"github.com/khunquant/khunquant/pkg/config"
	"github.com/khunquant/khunquant/pkg/exchanges"
)

func init() {
	exchanges.RegisterFactory("okx", func(cfg *config.Config) (exchanges.Exchange, error) {
		return NewOKXExchange(cfg)
	})
}
