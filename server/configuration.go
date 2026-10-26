package main

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

const (
	defaultMaxLifetimeDays = 7
	maxLifetimeDaysCap     = 30
)

// configuration mirrors settings_schema in plugin.json. The System Console can
// store a number setting as a JSON number or as a string, so it is decoded loosely.
type configuration struct {
	MaxLifetimeDays any
}

func (c configuration) maxLifetimeDays() (int, error) {
	var n int
	switch v := c.MaxLifetimeDays.(type) {
	case nil:
		return defaultMaxLifetimeDays, nil
	case float64:
		n = int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("MaxLifetimeDays must be a whole number, got %q", v)
		}
		n = parsed
	default:
		return 0, fmt.Errorf("MaxLifetimeDays has an unexpected type %T", v)
	}
	if n < 1 || n > maxLifetimeDaysCap {
		return 0, fmt.Errorf("MaxLifetimeDays must be between 1 and %d, got %d", maxLifetimeDaysCap, n)
	}
	return n, nil
}

type settings struct {
	maxLifetimeDays atomic.Int64
}

func (p *Plugin) OnConfigurationChange() error {
	var c configuration
	if err := p.API.LoadPluginConfiguration(&c); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	n, err := c.maxLifetimeDays()
	if err != nil {
		return err
	}
	p.settings.maxLifetimeDays.Store(int64(n))
	return nil
}
