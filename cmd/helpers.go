package cmd

import "github.com/namishh/vm-manager/internal/config"

// loadConfig es un helper para cargar la config desde cualquier comando
func loadConfig() (*config.Config, error) {
	return config.Load()
}
