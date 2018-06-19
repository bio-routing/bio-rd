package server

import "time"

type timer interface {
	Stop() bool
	Reset(d time.Duration) bool
}
