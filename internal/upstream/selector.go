package upstream

import (
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
)

var (
	rotatedAPIName string
	rotatedNext    time.Time
)

func InitSelector(conf config.ConfigUpstream) {
	switch conf.Mode {
	case config.RotatedUpstream:
		rotatedAPIName = conf.Upstream[rand.IntN(len(conf.Upstream))]
		rotatedNext = time.Now().Add(time.Duration(conf.RotatedInterval))
	}
}

func SelectAPI(conf config.ConfigUpstream) API {
	switch conf.Mode {
	case config.SingleUpstream:
		return New(conf.Upstream[0])
	case config.RandomUpstream:
		name := conf.Upstream[rand.IntN(len(conf.Upstream))]
		return New(name)
	case config.RotatedUpstream:
		// TODO update it in another goroutine
		if time.Now().After(rotatedNext) {
			rotatedAPIName = conf.Upstream[rand.IntN(len(conf.Upstream))]
			rotatedNext = rotatedNext.Add(time.Duration(conf.RotatedInterval))
		}
		return New(rotatedAPIName)
	}

	return nil
}
