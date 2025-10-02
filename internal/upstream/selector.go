package upstream

import (
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/debug"
)

var (
	rotatedName       string
	rotatedCycleEnded bool = false
)

func InitSelector(conf config.ConfigUpstream) {
	switch conf.Mode {
	case config.RotatedUpstream:
		rotatedName = randomCodename(conf.Upstream)
		newRotatedCycle(conf)
	}
}

func SelectAPI(conf config.ConfigUpstream) API {
	switch conf.Mode {
	case config.SingleUpstream:
		return New(conf.Upstream[0])
	case config.RandomUpstream:
		return New(randomCodename(conf.Upstream))
	case config.RotatedUpstream:
		if rotatedCycleEnded {
			rotatedCycleEnded = false
			newRotatedCycle(conf)
		}
		return New(rotatedName)
	}

	return nil
}

func randomCodename(list []string) string {
	return list[rand.IntN(len(list))]
}

func newRotatedCycle(conf config.ConfigUpstream) {
	debug.Logger.Println("new rotation cycle")
	time.AfterFunc(time.Duration(conf.RotatedInterval), func() {
		rotatedCycleEnded = true
		rotatedName = randomCodename(conf.Upstream)
	})
}
