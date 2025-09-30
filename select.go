package main

import (
	"math/rand/v2"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/upstream"
)

var (
	rotatedFrom     upstream.From
	rotatedInterval *time.Timer
)

func initAPI(conf config.ConfigUpstream) {
	switch conf.Type {
	case config.RotatedUpstream:
		rotatedFrom = conf.Upstream[rand.IntN(len(conf.Upstream))]
		rotatedInterval = time.NewTimer(time.Duration(conf.Interval))
	}
}

func getAPI(conf config.ConfigUpstream) upstream.API {
	switch conf.Type {
	case config.SingleUpstream:
		return upstream.Select(conf.Upstream[0])
	case config.RandomUpstream:
		from := conf.Upstream[rand.IntN(len(conf.Upstream))]
		return upstream.Select(from)
	case config.RotatedUpstream:
		select {
		case <-rotatedInterval.C:
			rotatedFrom = conf.Upstream[rand.IntN(len(conf.Upstream))]
			return upstream.Select(rotatedFrom)
		default:
			return upstream.Select(rotatedFrom)
		}
	}

	return nil
}
