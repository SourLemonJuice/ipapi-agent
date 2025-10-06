package upstream

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/SourLemonJuice/ipapi-agent/internal/config"
	"github.com/SourLemonJuice/ipapi-agent/internal/debug"
)

var (
	rotationMu        sync.RWMutex
	rotatedName       string
	rotatedCycleEnded bool
	rotationTimer     *time.Timer
)

func InitSelector(conf config.ConfigUpstream) error {
	if err := validateUpstreamConfig(conf); err != nil {
		return err
	}

	switch conf.Mode {
	case config.RotatedUpstream:
		if err := newRotatedCycle(conf); err != nil {
			return err
		}
	}

	return nil
}

func SelectAPI(conf config.ConfigUpstream) (API, error) {
	if len(conf.Upstream) == 0 {
		return nil, errors.New("upstream pool is empty")
	}

	switch conf.Mode {
	case config.SingleUpstream:
		return New(conf.Upstream[0])
	case config.RandomUpstream:
		codename, err := randomCodename(conf.Upstream)
		if err != nil {
			return nil, err
		}
		return New(codename)
	case config.RotatedUpstream:
		rotationMu.RLock()
		name := rotatedName
		cycleEnded := rotatedCycleEnded
		rotationMu.RUnlock()

		if cycleEnded {
			if err := newRotatedCycle(conf); err != nil {
				return nil, err
			}

			rotationMu.RLock()
			name = rotatedName
			rotationMu.RUnlock()
		}

		if name == "" {
			return nil, errors.New("rotated upstream not initialized")
		}

		return New(name)
	}

	return nil, errors.New("unsupported upstream mode")
}

func randomCodename(list []string) (string, error) {
	if len(list) == 0 {
		return "", errors.New("upstream pool is empty")
	}
	return list[rand.IntN(len(list))], nil
}

func newRotatedCycle(conf config.ConfigUpstream) error {
	name, err := randomCodename(conf.Upstream)
	if err != nil {
		return err
	}

	interval := time.Duration(conf.RotatedInterval)
	if interval <= 0 {
		return errors.New("rotated interval must be greater than zero")
	}

	rotationMu.Lock()
	if rotationTimer != nil {
		rotationTimer.Stop()
	}
	rotatedName = name
	rotatedCycleEnded = false
	rotationTimer = time.AfterFunc(interval, func() {
		rotationMu.Lock()
		rotatedCycleEnded = true
		rotationTimer = nil
		rotationMu.Unlock()
	})
	rotationMu.Unlock()

	debug.Logger.Println("new rotation cycle")
	return nil
}

func validateUpstreamConfig(conf config.ConfigUpstream) error {
	if len(conf.Upstream) == 0 {
		return errors.New("upstream pool is empty")
	}

	for _, codename := range conf.Upstream {
		if _, err := New(codename); err != nil {
			return fmt.Errorf("unsupported upstream %q: %w", codename, err)
		}
	}

	if conf.Mode == config.RotatedUpstream {
		if time.Duration(conf.RotatedInterval) <= 0 {
			return errors.New("rotated interval must be greater than zero")
		}
	}

	return nil
}
