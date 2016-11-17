package cfmysql

import "github.com/phayes/freeport"

//go:generate counterfeiter . PortFinder
type PortFinder interface {
	GetPort() int
}

type FreePortFinder struct {}

func (self *FreePortFinder) GetPort() int {
	return freeport.GetPort()
}
