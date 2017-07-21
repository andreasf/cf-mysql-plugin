package cfmysql

import "github.com/phayes/freeport"

//go:generate counterfeiter . PortFinder
type PortFinder interface {
	GetPort() int
}

func NewPortFinder() PortFinder {
	return new(portFinder)
}

type portFinder struct {}

func (self *portFinder) GetPort() int {
	return freeport.GetPort()
}
