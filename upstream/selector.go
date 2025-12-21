package upstream

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/config"
	C "github.com/SourLemonJuice/ipapi-agent/constant"
	"github.com/SourLemonJuice/ipapi-agent/debug"
)

var (
	rotateProvider string
)

func new(provider string) (API, error) {
	switch provider {
	case C.UpstreamProviderIpApiCom:
		return &ipApiCom{}, nil
	case C.UpstreamProviderIpinfoFree:
		return &ipinfoFree{}, nil
	case C.UpstreamProviderIpapiCo:
		return &ipapiCo{}, nil
	}

	return nil, fmt.Errorf("unknown upstream provider '%v'", provider)
}

func InitSelector(conf config.ConfigUpstream) {
	switch conf.Mode {
	case C.UpstreamModeRotate:
		go rotateRunner(conf)
	}
}

func rotateRunner(conf config.ConfigUpstream) {
	for {
		rotateProvider = randomProvider(conf.Pool)
		debug.Logger.Printf("New rotate_codename: %v", rotateProvider)
		time.Sleep(conf.RotateInterval)
	}
}

func SelectAPI(conf config.ConfigUpstream) (API, error) {
	prov := ""

	switch conf.Mode {
	case C.UpstreamModeSingle:
		prov = conf.Pool[0]
	case C.UpstreamModeRandom:
		prov = randomProvider(conf.Pool)
	case C.UpstreamModeRotate:
		prov = rotateProvider
	}

	api, err := new(prov)
	if err != nil {
		return nil, err
	}
	return api, nil
}

func randomProvider(list []string) string {
	return list[rand.IntN(len(list))]
}
