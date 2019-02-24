package types

import (
	"time"
)

type Events struct {
	Reason    string
	Message   string
	Source    string
	Type      string
	Count     int64
	Timestamp time.Time
}
