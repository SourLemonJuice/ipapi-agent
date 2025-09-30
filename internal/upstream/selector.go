package upstream

import "errors"

type From int

const (
	FromIpApiCom From = iota
	FromIpinfoFree
	FromIpapiCo
)

func Select(from From) API {
	switch from {
	case FromIpApiCom:
		return &ipApiCom{}
	case FromIpinfoFree:
		return &ipinfoFree{}
	case FromIpapiCo:
		return &ipapiCo{}
	}

	return nil
}

func ParseName(from string) (From, error) {
	switch from {
	case "ip-api.com":
		return FromIpApiCom, nil
	case "ipinfo-free":
		return FromIpinfoFree, nil
	case "ipapi.co":
		return FromIpapiCo, nil
	}

	return -1, errors.New("unknown upstream name")
}
