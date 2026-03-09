package cmd

import "github.com/Raulaxxo/vm-compose/internal/config"

// loadConfig es un helper para cargar la config desde cualquier comando
func loadConfig() (*config.Config, error) {
	return config.Load()
}
