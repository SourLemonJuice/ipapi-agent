package config

import (
	"errors"
	"fmt"
	"time"

	C "github.com/SourLemonJuice/ipapi-agent/constant"
)

type ConfigUpstream struct {
	Mode           string        `toml:"mode"`
	Pool           upstreamPool  `toml:"pool"`
	RotateInterval time.Duration `toml:"rotate_interval"`
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

func (upstream *ConfigUpstream) validate() error {
	switch upstream.Mode {
	case C.UpstreamModeSingle:
	case C.UpstreamModeRandom:
	case C.UpstreamModeRotate:
	default:
		return fmt.Errorf("upstream.mode has unknown type '%v'", upstream.Mode)
	}

	for _, v := range upstream.Pool {
		switch v {
		case C.UpstreamProviderIpApiCom:
		case C.UpstreamProviderIpinfoFree:
		case C.UpstreamProviderIpapiCo:
		default:
			return fmt.Errorf("upstream.pool has unknown provider '%v'", v)
		}
	}

	if upstream.RotateInterval <= 0 {
		return errors.New("upstream.rotate_interval has in not positive")
	}

	return nil
}
