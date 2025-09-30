package upstream

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
