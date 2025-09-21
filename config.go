package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	Listen        string    `toml:"listen"`
	ListenPort    uint16    `toml:"listen_port"`
	ResolveDomain bool      `toml:"resolve_domain"`
	Dev           configDev `toml:"dev"`
}

type configDev struct {
	Debug bool `toml:"debug"`
	Log   bool `toml:"log"`
}

func newConfig() config {
	return config{
		Listen:        "::",
		ListenPort:    8080,
		ResolveDomain: true,
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
