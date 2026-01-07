package config

type ConfigDomain struct {
	Enabled     bool     `toml:"enabled"`
	BlockSuffix []string `toml:"block_suffix"`
}

var DefaultDomain = ConfigDomain{
	Enabled:     true,
	BlockSuffix: nil,
}

func (domain *ConfigDomain) validate() error {
	// block some reserved TLDs
	// you may want to block .lan TLD with config file, because that's not a part of any standard.
	// https://en.wikipedia.org/wiki/Special-use_domain_name
	reservedSuffix := []string{"alt", "arpa", "invalid", "local", "localhost", "onion", "test", "internal"}
	domain.BlockSuffix = append(domain.BlockSuffix, reservedSuffix...)

	return nil
}
