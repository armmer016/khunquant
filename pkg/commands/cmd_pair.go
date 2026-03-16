package commands

import "context"

func pairCommand() Definition {
	return Definition{
		Name:        "pair",
		Description: "Request access to this bot",
		Usage:       "/pair",
		Handler: func(_ context.Context, req Request, _ *Runtime) error {
			return req.Reply("You are already authorized. ✅")
		},
	}
}
