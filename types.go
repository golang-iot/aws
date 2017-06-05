package aws

import (
	"time"
)

type Face struct {
	Device int64
	MaxAge int64
	MinAge int64
	Gender string
	GenderConf float64
	Smile bool
	SmileConf float64
	Emotions []string
	Created time.Time
}
