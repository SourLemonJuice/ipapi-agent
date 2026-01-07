package config

import (
	"errors"
	"time"
)

type ConfigDev struct {
	Debug           bool          `toml:"debug"`
	Log             bool          `toml:"log"`
	UpstreamTimeout time.Duration `toml:"upstream_timeout"`
}

func (dev *ConfigDev) validate() error {
	if dev.UpstreamTimeout <= 0 {
		return errors.New("upstream_timeout too short")
	}

	return nil
}
