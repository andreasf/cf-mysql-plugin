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

func NewPortWaiter(netWrapper NetWrapper) PortWaiter {
	return &portWaiter{
		NetWrapper: netWrapper,
	}
}

const SLEEP_TIME = 100

type portWaiter struct {
	NetWrapper NetWrapper
}

func (self *portWaiter) WaitUntilOpen(localPort int) {
	address := "127.0.0.1:" + strconv.Itoa(localPort)

	var conn net.Conn
	err := errors.New("")
	for err != nil {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		conn, err = self.NetWrapper.Dial("tcp", address)
	}
	self.NetWrapper.Close(conn)
}
