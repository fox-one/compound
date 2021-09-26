package config

import "github.com/fox-one/pkg/config"

// Load load config file
func Load(cfgFile string, cfg *Config) error {
	config.AutomaticLoadEnv("RINGS")
	if err := config.LoadYaml(cfgFile, cfg); err != nil {
		return err
	}

	defaultVote(cfg)
	return nil
}
