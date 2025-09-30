package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/SourLemonJuice/ipapi-agent/internal/upstream"
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
	Type     upstreamType            `toml:"type"`
	Upstream upstreamPool            `toml:"upstream"`
	Interval upstreamRotatedInterval `toml:"rotated_interval"`
}

type upstreamType int

const (
	SingleUpstream upstreamType = iota
	RandomUpstream
	RotatedUpstream
)

func (t *upstreamType) UnmarshalTOML(raw any) error {
	val, ok := raw.(string)
	if !ok {
		return errors.New("unknown value type")
	}

	switch val {
	case "single":
		*t = SingleUpstream
	case "random":
		*t = RandomUpstream
	case "rotated":
		*t = RotatedUpstream
	default:
		return errors.New("unknown upstream type")
	}

	return nil
}

type upstreamPool []upstream.From

func (pool *upstreamPool) UnmarshalTOML(raw any) error {
	valSingle, ok := raw.(string)
	if ok {
		from, err := upstream.ParseName(valSingle)
		if err != nil {
			return err
		}
		*pool = []upstream.From{from}
		return nil
	}

	valAnyArr, ok := raw.([]any)
	if ok {
		*pool = []upstream.From{} // init
		for _, v := range valAnyArr {
			valStr, ok := v.(string)
			if !ok {
				return errors.New("element not string")
			}
			from, err := upstream.ParseName(valStr)
			if err != nil {
				return err
			}
			*pool = append(*pool, from)
		}
		return nil
	}

	return errors.New("unknown value type")
}

type upstreamRotatedInterval time.Duration

func (interval *upstreamRotatedInterval) UnmarshalTOML(raw any) error {
	val, ok := raw.(string)
	if !ok {
		return errors.New("unknown value type")
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	if duration <= 0 {
		return errors.New("interval is in invalid range(<= 0)")
	}

	*interval = upstreamRotatedInterval(duration)
	return nil
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
			Type:     SingleUpstream,
			Upstream: []upstream.From{upstream.FromIpApiCom},
			Interval: upstreamRotatedInterval(time.Duration.Hours(24)),
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

	return nil
}
