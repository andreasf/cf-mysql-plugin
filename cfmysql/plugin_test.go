package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/onsi/gomega/gbytes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
)

var in *gbytes.Buffer
var out *gbytes.Buffer
var err *gbytes.Buffer
var cliConnection *pluginfakes.FakeCliConnection

var _ = Describe("Plugin", func() {
	BeforeEach(func() {
		in = gbytes.NewBuffer()
		out = gbytes.NewBuffer()
		err = gbytes.NewBuffer()
		cliConnection = new(pluginfakes.FakeCliConnection)
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
				sdkClient := new(cfmysqlfakes.FakeCfClient)
				plugin := MysqlPlugin{in, out, err, sdkClient}
				sdkClient.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)

				plugin.Run(cliConnection, []string{"mysql"})

				Expect(sdkClient.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(out).To(gbytes.Say("MySQL databases bound to an app:\n\ndatabase-a\ndatabase-b\n"))
				Expect(err).To(gbytes.Say(""))
			})
		})

		Context("With no databases available", func() {
			It("Tells the user that databases must be bound to a started app", func() {
				sdkClient := new(cfmysqlfakes.FakeCfClient)
				plugin := MysqlPlugin{in, out, err, sdkClient}
				sdkClient.GetMysqlServicesReturns([]MysqlService{}, nil)

				plugin.Run(cliConnection, []string{"mysql"})

				Expect(sdkClient.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(out).To(gbytes.Say(""))
				Expect(err).To(gbytes.Say("No MySQL databases available. Please bind your database services to a started app to make them available to 'cf mysql'."))
			})
		})
	})
})
