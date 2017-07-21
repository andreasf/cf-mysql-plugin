package cfmysql

import (
	"math/rand"
)

//go:generate counterfeiter . RandWrapper
type RandWrapper interface {
	Intn(n int) int
}

func NewRandWrapper() RandWrapper {
	return new(randWrapper)
}

type randWrapper struct{}

func (self *randWrapper) Intn(n int) int {
	return rand.Intn(n)
}
