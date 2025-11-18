package upstream

import (
	"math/rand/v2"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
)

func InitSelector(conf config.ConfigUpstream) {
}

func SelectAPI(conf config.ConfigUpstream) API {
	switch conf.Mode {
	case "single":
		return New(conf.Pool[0])
	case "random":
		return New(randomCodename(conf.Pool))
	case "rotated":
		// TODO
		panic("rotated upstream select mode didn't implemented")
	}

	return nil
}

func randomCodename(list []string) string {
	return list[rand.IntN(len(list))]
}
