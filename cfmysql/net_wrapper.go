package cfmysql

import "net"

//go:generate counterfeiter . Net
type Net interface {
	Dial(network, address string) (net.Conn, error)
	Close(conn net.Conn) error
}

type NetWrapper struct {}

func (self *NetWrapper) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

func (self *NetWrapper) Close(conn net.Conn) error {
	return conn.Close()
}
