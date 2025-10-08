package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Listen         string         `toml:"listen"`
	Port           uint16         `toml:"port"`
	TrustedProxies []string       `toml:"trusted_proxies"`
	Upstream       ConfigUpstream `toml:"upstream"`
	Resolve        ConfigResolve  `toml:"resolve"`
	Dev            ConfigDev      `toml:"dev"`
}

type ConfigUpstream struct {
	Mode            string        `toml:"mode"`
	Upstream        upstreamPool  `toml:"upstream"`
	RotatedInterval time.Duration `toml:"rotated_interval"`
}

type upstreamPool []string

// accept both a single string or a list of strings
func (pool *upstreamPool) UnmarshalTOML(raw any) error {
	valSingle, ok := raw.(string)
	if ok {
		*pool = []string{valSingle}
		return nil
	}

	valAnyArr, ok := raw.([]any)
	if ok {
		*pool = []string{} // init
		for _, v := range valAnyArr {
			valStr, ok := v.(string)
			if !ok {
				return errors.New("element not string")
			}
			*pool = append(*pool, valStr)
		}
		return nil
	}

	return errors.New("unknown value type")
}

type ConfigResolve struct {
	Domain   bool     `toml:"domain"`
	BlockTLD []string `toml:"block_tld"`
}

type ConfigDev struct {
	Debug bool `toml:"debug"`
	Log   bool `toml:"log"`
}

func New() Config {
	return Config{
		Listen:         "::",
		Port:           8080,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		Resolve: ConfigResolve{
			Domain:   true,
			BlockTLD: nil,
		},
		Upstream: ConfigUpstream{
			Mode:            "single",
			Upstream:        []string{"ip-api.com"},
			RotatedInterval: 24 * time.Hour,
		},
		Dev: ConfigDev{
			Debug: false,
			Log:   false,
		},
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

	err = c.valid()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) valid() error {
	switch c.Upstream.Mode {
	case "single":
	case "random":
	case "rotated":
	default:
		return errors.New("upstream.mode is unknown type")
	}

	if c.Upstream.RotatedInterval <= 0 {
		return errors.New("upstream.rotated_interval is in not positive")
	}

	return nil
}
