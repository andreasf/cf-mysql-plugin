package cfmysql

import (
	"time"
	"strconv"
	"errors"
	"net"
)

//go:generate counterfeiter . PortWaiter
type PortWaiter interface {
	WaitUntilOpen(localPort int)
}

const SLEEP_TIME = 100

type TcpPortWaiter struct {
	NetWrapper Net
}

func (self *TcpPortWaiter) WaitUntilOpen(localPort int) {
	address := "127.0.0.1:" + strconv.Itoa(localPort)

	var conn net.Conn
	err := errors.New("")
	for err != nil {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		conn, err = self.NetWrapper.Dial("tcp", address)
	}
	self.NetWrapper.Close(conn)
}

func NewPortWaiter(netWrapper Net) PortWaiter {
	return &TcpPortWaiter{
		NetWrapper: netWrapper,
	}
}
