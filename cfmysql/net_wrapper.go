package cfmysql

import "net"

//go:generate counterfeiter . NetWrapper
type NetWrapper interface {
	Dial(network, address string) (net.Conn, error)
	Close(conn net.Conn) error
}

func NewNetWrapper() NetWrapper {
	return new(netWrapper)
}

type netWrapper struct{}

func (self *netWrapper) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

func (self *netWrapper) Close(conn net.Conn) error {
	return conn.Close()
}
