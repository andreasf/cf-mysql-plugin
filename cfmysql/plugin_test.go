package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/onsi/gomega/gbytes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
)


var _ = Describe("Plugin", func() {
	var in *gbytes.Buffer
	var out *gbytes.Buffer
	var err *gbytes.Buffer
	var cliConnection *pluginfakes.FakeCliConnection
	var mysqlPlugin MysqlPlugin
	var apiClient *cfmysqlfakes.FakeApiClient

	BeforeEach(func() {
		in = gbytes.NewBuffer()
		out = gbytes.NewBuffer()
		err = gbytes.NewBuffer()
		cliConnection = new(pluginfakes.FakeCliConnection)
		apiClient = new(cfmysqlfakes.FakeApiClient)
		mysqlPlugin = MysqlPlugin{In: in, Out: out, Err: err, ApiClient: apiClient}
	})

	Context("When calling 'cf plugins'", func() {
		It("Shows the mysql plugin with version 1.0.0", func() {
			mysqlPlugin := NewPlugin()

			Expect(mysqlPlugin.GetMetadata().Name).To(Equal("mysql"))
			Expect(mysqlPlugin.GetMetadata().Version).To(Equal(plugin.VersionType{
				Major: 1,
				Minor: 0,
				Build: 0,
			}))
		})
	})

	Context("When calling 'cf mysql -h'", func() {
		It("Shows instructions for 'cf mysql'", func() {
			mysqlPlugin := NewPlugin()

			Expect(mysqlPlugin.GetMetadata().Commands).To(HaveLen(1))
			Expect(mysqlPlugin.GetMetadata().Commands[0].Name).To(Equal("mysql"))
		})
	})

	Context("When calling 'cf mysql' without arguments", func() {
		Context("With databases available", func() {
			var serviceA, serviceB MysqlService

			BeforeEach(func() {
				serviceA = MysqlService{
					Name: "database-a",
					Hostname: "database-a.host",
					Port: "123",
					DbName: "dbname-a",
				}
				serviceB = MysqlService{
					Name: "database-b",
					Hostname: "database-b.host",
					Port: "234",
					DbName: "dbname-b",
				}
			})

			It("Lists the available MySQL databases", func() {
				apiClient.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)

				mysqlPlugin.Run(cliConnection, []string{"mysql"})

				Expect(apiClient.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(out).To(gbytes.Say("MySQL databases bound to an app:\n\ndatabase-a\ndatabase-b\n"))
				Expect(err).To(gbytes.Say(""))
				Expect(mysqlPlugin.ExitCode).To(Equal(0))
			})
		})

		Context("With no databases available", func() {
			It("Tells the user that databases must be bound to a started app", func() {
				apiClient.GetMysqlServicesReturns([]MysqlService{}, nil)

				mysqlPlugin.Run(cliConnection, []string{"mysql"})

				Expect(apiClient.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(out).To(gbytes.Say(""))
				Expect(err).To(gbytes.Say("No MySQL databases available. Please bind your database services to a started app to make them available to 'cf mysql'."))
				Expect(mysqlPlugin.ExitCode).To(Equal(0))
			})
		})

		Context("With failing API calls", func() {
			It("Shows an error message", func() {
				apiClient.GetMysqlServicesReturns(nil, fmt.Errorf("foo"))

				mysqlPlugin.Run(cliConnection, []string{"mysql"})

				Expect(apiClient.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(out).To(gbytes.Say(""))
				Expect(err).To(gbytes.Say("Unable to retrieve services: foo\n"))
				Expect(mysqlPlugin.ExitCode).To(Equal(1))
			})
		})
	})
})
