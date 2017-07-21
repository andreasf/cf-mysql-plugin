package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"net"
	"errors"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/netfakes"
)

var _ = Describe("PortWaiter", func() {
	var netWrapper *cfmysqlfakes.FakeNetWrapper
	var portWaiter *TcpPortWaiter

	const SUCCEED_AFTER_TRIES = 5
	dialCount := 0
	mockDial := func(network, address string) (net.Conn, error) {
		if dialCount < SUCCEED_AFTER_TRIES - 1 {
			dialCount++
			return nil, errors.New("GURU MEDITATION")
		}

		return new(netfakes.FakeConn), nil
	}

	BeforeEach(func() {
		netWrapper = new(cfmysqlfakes.FakeNetWrapper)
		portWaiter = &TcpPortWaiter{
			NetWrapper: netWrapper,
		}
	})

	It("Waits until the port is open", func() {
		netWrapper.DialStub = mockDial

		portWaiter.WaitUntilOpen(523)

		Expect(netWrapper.DialCallCount()).To(Equal(SUCCEED_AFTER_TRIES))
	})

	It("Closes the connection", func() {
		mockConn := new(netfakes.FakeConn)
		netWrapper.DialReturns(mockConn, nil)

		portWaiter.WaitUntilOpen(523)

		Expect(netWrapper.CloseCallCount()).To(Equal(1))
		Expect(netWrapper.CloseArgsForCall(0)).To(Equal(mockConn))
		Expect(netWrapper.DialCallCount()).To(Equal(1))
	})
})
