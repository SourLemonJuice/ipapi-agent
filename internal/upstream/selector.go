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
	case "rotated":
		rotatedName = randomCodename(conf.Pool)
		newRotatedCycle(conf)
	}
}

func SelectAPI(conf config.ConfigUpstream) API {
	switch conf.Mode {
	case "single":
		return New(conf.Pool[0])
	case "random":
		return New(randomCodename(conf.Pool))
	case "rotated":
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
		rotatedName = randomCodename(conf.Pool)
	})
}
