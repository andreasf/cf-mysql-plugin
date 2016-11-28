package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"errors"
	"os"
)

var _ = Describe("MysqlRunner", func() {
	Context("RunMysql", func() {
		var exec *cfmysqlfakes.FakeExec
		var runner MysqlClientRunner

		BeforeEach(func() {
			exec = new(cfmysqlfakes.FakeExec)
			runner = MysqlClientRunner{
				ExecWrapper: exec,
			}
		})

		Context("When mysql is not in PATH", func() {
			It("Returns an error", func() {
				exec.LookPathReturns("", errors.New("PC LOAD LETTER"))

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password")

				Expect(err).To(Equal(errors.New("'mysql' client not found in PATH")))
			})
		})

		Context("When Run returns an error", func() {
			It("Forwards the error", func() {
				exec.LookPathReturns("/path/to/mysql", nil)
				exec.RunReturns(errors.New("PC LOAD LETTER"))

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password")

				Expect(err).To(Equal(errors.New("Error running mysql client: PC LOAD LETTER")))
			})
		})

		Context("When mysql is in PATH", func() {
			It("Calls mysql with the right arguments", func() {
				exec.LookPathReturns("/path/to/mysql", nil)

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password")

				Expect(err).To(BeNil())
				Expect(exec.LookPathCallCount()).To(Equal(1))
				Expect(exec.RunCallCount()).To(Equal(1))

				cmd := exec.RunArgsForCall(0)
				Expect(cmd.Path).To(Equal("/path/to/mysql"))
				Expect(cmd.Args).To(Equal([]string{"/path/to/mysql", "-u", "username", "-ppassword", "-h", "hostname", "-P", "42", "dbname"}))
				Expect(cmd.Stdin).To(Equal(os.Stdin))
				Expect(cmd.Stdout).To(Equal(os.Stdout))
				Expect(cmd.Stderr).To(Equal(os.Stderr))
			})
		})

		Context("When mysql is in PATH and additional arguments are passed", func() {
			It("Calls mysql with the right arguments", func() {
				exec.LookPathReturns("/path/to/mysql", nil)

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password", "--foo", "bar", "--baz")

				Expect(err).To(BeNil())
				Expect(exec.LookPathCallCount()).To(Equal(1))
				Expect(exec.RunCallCount()).To(Equal(1))

				cmd := exec.RunArgsForCall(0)
				Expect(cmd.Path).To(Equal("/path/to/mysql"))
				Expect(cmd.Args).To(Equal([]string{"/path/to/mysql", "-u", "username", "-ppassword", "-h", "hostname", "-P", "42", "--foo", "bar", "--baz", "dbname"}))
				Expect(cmd.Stdin).To(Equal(os.Stdin))
				Expect(cmd.Stdout).To(Equal(os.Stdout))
				Expect(cmd.Stderr).To(Equal(os.Stderr))
			})
		})
	})
})
