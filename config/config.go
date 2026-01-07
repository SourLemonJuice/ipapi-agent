package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Listen         string         `toml:"listen"`
	Port           uint16         `toml:"port"`
	TrustedProxies []string       `toml:"trusted_proxies"`
	Upstream       ConfigUpstream `toml:"upstream"`
	Domain         ConfigDomain   `toml:"domain"`
	Dev            ConfigDev      `toml:"dev"`
}

func Default() Config {
	return Config{
		Listen:         "::",
		Port:           8080,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		Domain:         DefaultDomain,
		Upstream:       DefaultUpstream,
		Dev:            DefaultDev,
	}
}

func (c *Config) DecodeFile(path string) error {
	var err error

	md, err := toml.DecodeFile(path, c)
	if err != nil {
		return fmt.Errorf("decode TOML error: %w", err)
	}

	unknowns := md.Undecoded()
	if len(unknowns) != 0 {
		return fmt.Errorf("invalid TOML keys: %v", unknowns)
	}

	err = c.validate()
	if err != nil {
		return err
	}

	return nil
}

func (conf *Config) validate() error {
	err := conf.Upstream.validate()
	if err != nil {
		return err
	}

	err = conf.Domain.validate()
	if err != nil {
		return err
	}

	err = conf.Dev.validate()
	if err != nil {
		return err
	}

	return nil
}
