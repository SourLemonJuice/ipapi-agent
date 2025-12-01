package upstream

import (
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/debug"
)

var (
	rotateCodename string
)

func InitSelector(conf config.ConfigUpstream) {
	switch conf.Mode {
	case config.ModeRotate:
		go rotate(conf)
	}
}

func SelectAPI(conf config.ConfigUpstream) API {
	switch conf.Mode {
	case config.ModeSingle:
		return New(conf.Pool[0])
	case config.ModeRandom:
		return New(randomCodename(conf.Pool))
	case config.ModeRotate:
		return New(rotateCodename)
	}

	return nil
}

func randomCodename(list []string) string {
	return list[rand.IntN(len(list))]
}

func rotate(conf config.ConfigUpstream) {
	for {
		rotateCodename = randomCodename(conf.Pool)
		debug.Logger.Printf("New rotate_codename: %v", rotateCodename)
		time.Sleep(conf.RotateInterval)
	}
}
