package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	Listen         string        `toml:"listen"`
	Port           uint16        `toml:"port"`
	TrustedProxies []string      `toml:"trusted_proxies"`
	Resolve        configResolve `toml:"resolve"`
	Dev            configDev     `toml:"dev"`
}

type configResolve struct {
	Domain       bool     `toml:"domain"`
	TLDBlockList []string `toml:"tld_blocklist"`
}

type configDev struct {
	Debug bool `toml:"debug"`
	Log   bool `toml:"log"`
}

func newConfig() config {
	return config{
		Listen:         "::",
		Port:           8080,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		Resolve: configResolve{
			Domain:       true,
			TLDBlockList: nil,
		},
		Dev: configDev{
			Debug: false,
			Log:   false,
		},
	}
}

func (c *config) decodeFile(path string) error {
	var err error

	md, err := toml.DecodeFile(path, c)
	if err != nil {
		return fmt.Errorf("decode TOML error: %w", err)
	}

	unknowns := md.Undecoded()
	if len(unknowns) != 0 {
		return fmt.Errorf("invalid TOML keys: %v", unknowns)
	}

	return nil
}
