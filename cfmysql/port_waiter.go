package cfmysql

import "time"

//go:generate counterfeiter . PortWaiter
type PortWaiter interface {
	WaitUntilOpen(localPort int)
}

type TcpPortWaiter struct{}

func (self *TcpPortWaiter) WaitUntilOpen(localPort int) {
	time.Sleep(5 * time.Second)
}
