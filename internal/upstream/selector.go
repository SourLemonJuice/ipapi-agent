package upstream

import (
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/debug"
)

var (
	rotate_codename string
)

func InitSelector(conf config.ConfigUpstream) {
	switch conf.Mode {
	case "rotate":
		go rotate(conf)
	}
}

func SelectAPI(conf config.ConfigUpstream) API {
	switch conf.Mode {
	case "single":
		return New(conf.Pool[0])
	case "random":
		return New(randomCodename(conf.Pool))
	case "rotate":
		return New(rotate_codename)
	}

	return nil
}

func randomCodename(list []string) string {
	return list[rand.IntN(len(list))]
}

func rotate(conf config.ConfigUpstream) {
	for {
		rotate_codename = randomCodename(conf.Pool)
		debug.Logger.Printf("New rotate_codename: %v", rotate_codename)
		time.Sleep(conf.RotateInterval)
	}
}
