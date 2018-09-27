package cfmysql_test

import (
	"errors"
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("MysqlRunner", func() {
	Context("RunMysql", func() {
		var exec *cfmysqlfakes.FakeExecWrapper
		var ioutilWrapper *cfmysqlfakes.FakeIoUtilWrapper
		var osWrapper *cfmysqlfakes.FakeOsWrapper
		var runner MysqlRunner

		BeforeEach(func() {
			exec = new(cfmysqlfakes.FakeExecWrapper)
			ioutilWrapper = new(cfmysqlfakes.FakeIoUtilWrapper)
			osWrapper = new(cfmysqlfakes.FakeOsWrapper)
			runner = NewMysqlRunner(exec, ioutilWrapper, osWrapper)
		})

		Context("When mysql is not in PATH", func() {
			It("Returns an error", func() {
				exec.LookPathReturns("", errors.New("PC LOAD LETTER"))

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password", "")

				Expect(err).To(Equal(errors.New("'mysql' client not found in PATH")))
				Expect(exec.LookPathArgsForCall(0)).To(Equal("mysql"))
			})
		})

		Context("When Run returns an error", func() {
			It("Forwards the error", func() {
				exec.LookPathReturns("/path/to/mysql", nil)
				exec.RunReturns(errors.New("PC LOAD LETTER"))

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password", "")

				Expect(err).To(Equal(errors.New("error running mysql client: PC LOAD LETTER")))
			})
		})

		Context("When mysql is in PATH", func() {
			It("Calls mysql with the right arguments", func() {
				exec.LookPathReturns("/path/to/mysql", nil)

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password", "")

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

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password", "", "--foo", "bar", "--baz")

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

		Context("When mysql is in PATH and a TLS CA certificate is part of the service credentials", func() {
			It("Stores the cert in a temp file and calls mysql with --ssl-ca=path", func() {
				exec.LookPathReturns("/path/to/mysql", nil)
				tempFile := new(os.File)
				ioutilWrapper.TempFileReturns(tempFile, nil)
				osWrapper.NameReturns("/path/to/cert.pem")

				err := runner.RunMysql("hostname", 42, "dbname", "username", "password", "cert-content", "--foo", "bar", "--baz")

				Expect(err).To(BeNil())
				Expect(exec.LookPathCallCount()).To(Equal(1))
				Expect(ioutilWrapper.TempFileCallCount()).To(Equal(1))
				Expect(osWrapper.WriteStringCallCount()).To(Equal(1))
				Expect(osWrapper.NameCallCount()).To(Equal(1))
				Expect(exec.RunCallCount()).To(Equal(1))
				Expect(osWrapper.RemoveCallCount()).To(Equal(1))

				tempFileDir, tempFilePattern := ioutilWrapper.TempFileArgsForCall(0)
				Expect(tempFileDir).To(Equal(""))
				Expect(tempFilePattern).To(Equal("mysql-ca-cert.pem"))

				writeStringFile, writeStringString := osWrapper.WriteStringArgsForCall(0)
				Expect(writeStringFile).To(BeIdenticalTo(tempFile))
				Expect(writeStringString).To(Equal("cert-content"))

				cmd := exec.RunArgsForCall(0)
				Expect(cmd.Path).To(Equal("/path/to/mysql"))
				Expect(cmd.Args).To(Equal([]string{"/path/to/mysql", "-u", "username", "-ppassword", "-h", "hostname", "-P", "42", "--ssl-ca=/path/to/cert.pem", "--foo", "bar", "--baz", "dbname"}))
				Expect(cmd.Stdin).To(Equal(os.Stdin))
				Expect(cmd.Stdout).To(Equal(os.Stdout))
				Expect(cmd.Stderr).To(Equal(os.Stderr))

				removePath := osWrapper.RemoveArgsForCall(0)
				Expect(removePath).To(Equal("/path/to/cert.pem"))
			})
		})
	})

	Context("RunMysqlDump", func() {
		var exec *cfmysqlfakes.FakeExecWrapper
		var ioutil *cfmysqlfakes.FakeIoUtilWrapper
		var osWrapper *cfmysqlfakes.FakeOsWrapper
		var runner MysqlRunner

		BeforeEach(func() {
			exec = new(cfmysqlfakes.FakeExecWrapper)
			ioutil = new(cfmysqlfakes.FakeIoUtilWrapper)
			osWrapper = new(cfmysqlfakes.FakeOsWrapper)
			runner = NewMysqlRunner(exec, ioutil, osWrapper)
		})

		Context("When mysqldump is not in PATH", func() {
			It("Returns an error", func() {
				exec.LookPathReturns("", errors.New("PC LOAD LETTER"))

				err := runner.RunMysqlDump("hostname", 42, "dbname", "username", "password")

				Expect(err).To(Equal(errors.New("'mysqldump' not found in PATH")))
				Expect(exec.LookPathArgsForCall(0)).To(Equal("mysqldump"))
			})
		})

		Context("When Run returns an error", func() {
			It("Forwards the error", func() {
				exec.LookPathReturns("/path/to/mysqldump", nil)
				exec.RunReturns(errors.New("PC LOAD LETTER"))

				err := runner.RunMysqlDump("hostname", 42, "dbname", "username", "password")

				Expect(err).To(Equal(errors.New("error running mysqldump: PC LOAD LETTER")))
			})
		})

		Context("When mysqldump is in PATH", func() {
			It("Calls mysqldump with the right arguments", func() {
				exec.LookPathReturns("/path/to/mysqldump", nil)

				err := runner.RunMysqlDump("hostname", 42, "dbname", "username", "password")

				Expect(err).To(BeNil())
				Expect(exec.LookPathCallCount()).To(Equal(1))
				Expect(exec.RunCallCount()).To(Equal(1))

				cmd := exec.RunArgsForCall(0)
				Expect(cmd.Path).To(Equal("/path/to/mysqldump"))
				Expect(cmd.Args).To(Equal([]string{"/path/to/mysqldump", "-u", "username", "-ppassword", "-h", "hostname", "-P", "42", "dbname"}))
				Expect(cmd.Stdin).To(Equal(os.Stdin))
				Expect(cmd.Stdout).To(Equal(os.Stdout))
				Expect(cmd.Stderr).To(Equal(os.Stderr))
			})
		})

		Context("When mysqldump is in PATH and additional arguments are passed", func() {
			It("Calls mysqldump with the right arguments", func() {
				exec.LookPathReturns("/path/to/mysqldump", nil)

				err := runner.RunMysqlDump("hostname", 42, "dbname", "username", "password", "table1", "table2", "--foo", "bar", "--baz")

				Expect(err).To(BeNil())
				Expect(exec.LookPathCallCount()).To(Equal(1))
				Expect(exec.RunCallCount()).To(Equal(1))

				cmd := exec.RunArgsForCall(0)
				Expect(cmd.Path).To(Equal("/path/to/mysqldump"))
				Expect(cmd.Args).To(Equal([]string{"/path/to/mysqldump", "-u", "username", "-ppassword", "-h", "hostname", "-P", "42", "--foo", "bar", "--baz", "dbname", "table1", "table2"}))
				Expect(cmd.Stdin).To(Equal(os.Stdin))
				Expect(cmd.Stdout).To(Equal(os.Stdout))
				Expect(cmd.Stderr).To(Equal(os.Stderr))
			})
		})
	})
})
