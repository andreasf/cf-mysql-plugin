package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
)

var _ = Describe("SshRunner", func() {
	var cliConnection *pluginfakes.FakeCliConnection
	var sshRunner SshRunner
	service := MysqlService{
		Name: "database-a",
		AppName: "",
		Hostname: "database-a.host",
		Port: "3306",
		DbName: "dbname-a",
		Username: "username-a",
		Password: "password-a",
	}

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		sshRunner = NewSshRunner()
	})

	Context("When opening the tunnel", func() {
		It("Runs 'cf ssh'", func() {
			sshRunner.OpenSshTunnel(cliConnection, service, "app-name", 4242)

			Expect(cliConnection.CliCommandCallCount()).To(Equal(1))
			calledArgs := cliConnection.CliCommandArgsForCall(0)
			Expect(calledArgs).To(Equal([]string{"ssh", "app-name", "-N", "-L", "4242:database-a.host:3306"}))
		})
	})

	Context("When 'cf ssh' returns an error", func() {
		It("panics", func() {
			defer func() {
				if thePanic := recover(); thePanic != nil {
					Expect(thePanic).To(Equal(errors.New("SSH tunnel failed: PC LOAD LETTER")))
				} else {
					Fail("Expected function to panic")
				}
			}()
			cliConnection.CliCommandWithoutTerminalOutputStub = nil
			cliConnection.CliCommandReturns(nil, errors.New("PC LOAD LETTER"))

			sshRunner.OpenSshTunnel(cliConnection, service, "app-name", 4242)
		})
	})
})
