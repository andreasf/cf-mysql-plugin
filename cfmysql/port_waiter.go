package cfmysql

import (
	"errors"
	"net"
	"strconv"
	"time"
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

const SleepTime = 100

type portWaiter struct {
	NetWrapper NetWrapper
}

func (self *portWaiter) WaitUntilOpen(localPort int) {
	address := "127.0.0.1:" + strconv.Itoa(localPort)

	var conn net.Conn
	err := errors.New("")
	for err != nil {
		time.Sleep(SleepTime * time.Millisecond)
		conn, err = self.NetWrapper.Dial("tcp", address)
	}
	self.NetWrapper.Close(conn)
}
