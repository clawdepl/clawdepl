package cmd

import (
	"fmt"

	"github.com/clawdepl/clawdepl/internal/config"
)

func requireLogin() error {
	if HasUnsafeToken() || config.IsLoggedIn() {
		return nil
	}
	return fmt.Errorf("not logged in. run 'clawdepl login' first")
}
