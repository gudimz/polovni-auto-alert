package polovniauto

import "github.com/kelseyhightower/envconfig"

type Config struct {
	PageLimit int `envconfig:"PAGE_LIMIT" default:"9999"`
}

func NewConfig() *Config {
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	return cfg
}
