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
	Upstream       configUpstream `toml:"upstream"`
	Resolve        configResolve  `toml:"resolve"`
	Dev            configDev      `toml:"dev"`
}

type configUpstream struct {
	Type     UpstreamType     `toml:"type"`     // to UpstreamType
	Upstream upstreamPool     `toml:"upstream"` // when any type
	Interval upstreamInterval `toml:"interval"` // when UpstreamRotation
}

type UpstreamType int

const (
	UpstreamSingle UpstreamType = iota
	UpstreamMultiple
	UpstreamRotation
	UpstreamSchedule
)

func (t *UpstreamType) UnmarshalTOML(raw any) error {
	val, ok := raw.(string)
	if !ok {
		return errors.New("unknown value type")
	}

	switch val {
	case "single":
		*t = UpstreamSingle
	case "multiple":
		*t = UpstreamMultiple
	case "rotation":
		*t = UpstreamRotation
	case "schedule":
		*t = UpstreamSchedule
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

	valArr, ok := raw.([]string)
	if ok {
		*pool = []upstream.From{} // init
		for _, v := range valArr {
			from, err := upstream.ParseName(v)
			if err != nil {
				return err
			}
			*pool = append(*pool, from)
		}
		return nil
	}

	return errors.New("unknown value type")
}

type upstreamInterval time.Duration

func (interval *upstreamInterval) UnmarshalTOML(raw any) error {
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

	*interval = upstreamInterval(duration)
	return nil
}

type configResolve struct {
	Domain   bool     `toml:"domain"`
	BlockTLD []string `toml:"block_tld"`
}

type configDev struct {
	Debug bool `toml:"debug"`
	Log   bool `toml:"log"`
}

func New() Config {
	return Config{
		Listen:         "::",
		Port:           8080,
		TrustedProxies: []string{"127.0.0.1", "::1"},
		Resolve: configResolve{
			Domain:   true,
			BlockTLD: nil,
		},
		Upstream: configUpstream{
			Type:     UpstreamSingle,
			Upstream: []upstream.From{upstream.FromIpApiCom},
			Interval: upstreamInterval(time.Duration.Hours(24)),
		},
		Dev: configDev{
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
