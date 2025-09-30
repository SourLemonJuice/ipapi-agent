package main

import (
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/upstream"
)

var (
	rotatedFrom upstream.From
	rotatedNext time.Time
)

func initAPI(conf config.ConfigUpstream) {
	switch conf.Mode {
	case config.RotatedUpstream:
		rotatedFrom = conf.Upstream[rand.IntN(len(conf.Upstream))]
		rotatedNext = time.Now().Add(time.Duration(conf.RotatedInterval))
	}
}

func getAPI(conf config.ConfigUpstream) upstream.API {
	switch conf.Mode {
	case config.SingleUpstream:
		return upstream.Select(conf.Upstream[0])
	case config.RandomUpstream:
		from := conf.Upstream[rand.IntN(len(conf.Upstream))]
		return upstream.Select(from)
	case config.RotatedUpstream:
		if time.Now().After(rotatedNext) {
			rotatedFrom = conf.Upstream[rand.IntN(len(conf.Upstream))]
			rotatedNext = rotatedNext.Add(time.Duration(conf.RotatedInterval))
		}
		return upstream.Select(rotatedFrom)
	}

	return nil
}
