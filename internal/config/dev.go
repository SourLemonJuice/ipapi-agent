package config

type ConfigDev struct {
	Debug bool `toml:"debug"`
	Log   bool `toml:"log"`
}

func (dev *ConfigDev) validate() error {
	return nil
}
