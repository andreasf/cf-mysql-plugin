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
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/cf/errors"
)

var _ = Describe("Plugin", func() {
	var appList []plugin_models.GetAppsModel
	usage := "cf mysql - Connect to a MySQL database service\n\nUSAGE:\n   Open a mysql client to a database:\n   cf mysql <service-name> [mysql args...]\n\n\ncf mysqldump - Dump a MySQL database\n\nUSAGE:\n   Dump all tables in a database:\n   cf mysqldump <service-name> [mysqldump args...]\n   Dump specific tables in a database:\n   cf mysqldump <service-name> [tables...] [mysqldump args...]\n"

	BeforeEach(func() {
		appList = []plugin_models.GetAppsModel{
			{
				Name: "app-name-1",
			},
			{
				Name: "app-name-2",
			},
		}
	})

	Context("When calling 'cf plugins'", func() {
		It("Shows the mysql plugin with the current version", func() {
			mysqlPlugin, _ := NewPluginAndMocks()

			Expect(mysqlPlugin.GetMetadata().Name).To(Equal("mysql"))
			Expect(mysqlPlugin.GetMetadata().Version).To(Equal(plugin.VersionType{
				Major: 2,
				Minor: 0,
				Build: 0,
			}))
		})
	})

	Context("When calling 'cf mysql -h'", func() {
		It("Shows instructions for 'cf mysql'", func() {
			mysqlPlugin, _ := NewPluginAndMocks()

			Expect(mysqlPlugin.GetMetadata().Commands).To(HaveLen(2))
			Expect(mysqlPlugin.GetMetadata().Commands[0].Name).To(Equal("mysql"))
		})
	})

	Context("When calling 'cf mysql' without arguments", func() {
		It("Prints usage instructions to STDERR and exits with 1", func() {
			mysqlPlugin, mocks := NewPluginAndMocks()

			mysqlPlugin.Run(mocks.CliConnection, []string{"mysql"})

			Expect(mocks.Out).To(gbytes.Say(""))
			Expect(string(mocks.Err.Contents())).To(Equal(usage))
			Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
		})
	})

	Context("When calling 'cf mysql db-name'", func() {
		var serviceA MysqlService

		BeforeEach(func() {
			serviceA = MysqlService{
				Name:     "database-a",
				Hostname: "database-a.host",
				Port:     "123",
				DbName:   "dbname-a",
				Username: "username",
				Password: "password",
			}
		})

		Context("When the database is available", func() {
			It("Opens an SSH tunnel through a started app", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(serviceA, nil)
				mocks.CfService.GetStartedAppsReturns(appList, nil)
				mocks.PortFinder.GetPortReturns(2342)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.CfService.GetServiceCallCount()).To(Equal(1))
				calledCliConnection, calledName := mocks.CfService.GetServiceArgsForCall(0)
				Expect(calledName).To(Equal("database-a"))
				Expect(calledCliConnection).To(Equal(mocks.CliConnection))

				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
				Expect(mocks.CfService.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppList, localPort := mocks.CfService.OpenSshTunnelArgsForCall(0)
				Expect(calledCliConnection).To(Equal(mocks.CliConnection))
				Expect(calledService).To(Equal(serviceA))
				Expect(calledAppList).To(Equal(appList))
				Expect(localPort).To(Equal(2342))
			})

			It("Opens a MySQL client connecting through the tunnel", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(serviceA, nil)
				mocks.CfService.GetStartedAppsReturns(appList, nil)
				mocks.PortFinder.GetPortReturns(2342)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
				Expect(mocks.MysqlRunner.RunMysqlCallCount()).To(Equal(1))

				hostname, port, dbName, username, password, _ := mocks.MysqlRunner.RunMysqlArgsForCall(0)
				Expect(hostname).To(Equal("127.0.0.1"))
				Expect(port).To(Equal(2342))
				Expect(dbName).To(Equal(serviceA.DbName))
				Expect(username).To(Equal(serviceA.Username))
				Expect(password).To(Equal(serviceA.Password))
			})

			Context("When passing additional arguments", func() {
				It("Passes the arguments to mysql", func() {
					mysqlPlugin, mocks := NewPluginAndMocks()

					mocks.CfService.GetServiceReturns(serviceA, nil)
					mocks.CfService.GetStartedAppsReturns(appList, nil)
					mocks.PortFinder.GetPortReturns(2342)

					mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a", "--foo", "bar", "--baz"})

					Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
					Expect(mocks.MysqlRunner.RunMysqlCallCount()).To(Equal(1))

					hostname, port, dbName, username, password, args := mocks.MysqlRunner.RunMysqlArgsForCall(0)
					Expect(hostname).To(Equal("127.0.0.1"))
					Expect(port).To(Equal(2342))
					Expect(dbName).To(Equal(serviceA.DbName))
					Expect(username).To(Equal(serviceA.Username))
					Expect(password).To(Equal(serviceA.Password))
					Expect(args).To(Equal([]string{"--foo", "bar", "--baz"}))
				})
			})
		})

		Context("When a service key cannot be retrieved", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(MysqlService{}, errors.New("database not found"))

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "db-name"})

				Expect(mocks.CfService.GetServiceCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("^FAILED\nUnable to retrieve service credentials: database not found\n$"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})

		Context("When there are no started apps", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(serviceA, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.CfService.GetServiceCallCount()).To(Equal(1))
				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say("^$"))
				Expect(mocks.Err).To(gbytes.Say("^FAILED\nUnable to connect to 'database-a': no started apps in current space\n$"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})

		Context("When GetStartedApps returns an error", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(serviceA, nil)
				mocks.CfService.GetStartedAppsReturns(nil, fmt.Errorf("PC LOAD LETTER"))

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.CfService.GetServiceCallCount()).To(Equal(1))
				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("^FAILED\nUnable to retrieve started apps: PC LOAD LETTER\n$"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})

	})

	Context("When calling 'cf mysqldump -h'", func() {
		It("Shows instructions for 'cf mysqldump'", func() {
			mysqlPlugin, _ := NewPluginAndMocks()

			Expect(mysqlPlugin.GetMetadata().Commands).To(HaveLen(2))
			Expect(mysqlPlugin.GetMetadata().Commands[1].Name).To(Equal("mysqldump"))
		})
	})

	Context("When calling 'cf mysqldump' without arguments", func() {
		It("Prints usage information to STDERR and exits with 1", func() {
			mysqlPlugin, mocks := NewPluginAndMocks()

			mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump"})

			Expect(mocks.Out).To(gbytes.Say(""))
			Expect(string(mocks.Err.Contents())).To(Equal(usage))
			Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
		})
	})

	Context("When calling 'cf mysqldump db-name'", func() {
		var serviceA MysqlService

		BeforeEach(func() {
			serviceA = MysqlService{
				Name:     "database-a",
				Hostname: "database-a.host",
				Port:     "123",
				DbName:   "dbname-a",
				Username: "username",
				Password: "password",
			}
		})

		Context("When the database is available", func() {
			var app1 plugin_models.GetAppsModel
			var app2 plugin_models.GetAppsModel

			BeforeEach(func() {
				app1 = plugin_models.GetAppsModel{
					Name: "app-name-1",
				}
				app2 = plugin_models.GetAppsModel{
					Name: "app-name-2",
				}
			})

			It("Opens an SSH tunnel through a started app", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(serviceA, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app1, app2}, nil)
				mocks.PortFinder.GetPortReturns(2342)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump", "database-a"})

				Expect(mocks.CfService.GetServiceCallCount()).To(Equal(1))
				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
				Expect(mocks.CfService.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppList, localPort := mocks.CfService.OpenSshTunnelArgsForCall(0)
				Expect(calledCliConnection).To(Equal(mocks.CliConnection))
				Expect(calledService).To(Equal(serviceA))
				Expect(calledAppList).To(Equal(appList))
				Expect(localPort).To(Equal(2342))
			})

			It("Opens mysqldump connecting through the tunnel", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetServiceReturns(serviceA, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app1}, nil)
				mocks.PortFinder.GetPortReturns(2342)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump", "database-a"})

				Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
				Expect(mocks.MysqlRunner.RunMysqlDumpCallCount()).To(Equal(1))

				hostname, port, dbName, username, password, _ := mocks.MysqlRunner.RunMysqlDumpArgsForCall(0)
				Expect(hostname).To(Equal("127.0.0.1"))
				Expect(port).To(Equal(2342))
				Expect(dbName).To(Equal(serviceA.DbName))
				Expect(username).To(Equal(serviceA.Username))
				Expect(password).To(Equal(serviceA.Password))
			})
		})
	})

	Context("When uninstalling the plugin", func() {
		It("Does not give any output or call the API", func() {
			mysqlPlugin, mocks := NewPluginAndMocks()

			mysqlPlugin.Run(mocks.CliConnection, []string{"CLI-MESSAGE-UNINSTALL"})

			Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(0))
			Expect(mocks.Out).To(gbytes.Say("^$"))
			Expect(mocks.Err).To(gbytes.Say("^$"))
			Expect(mysqlPlugin.GetExitCode()).To(Equal(0))
		})
	})
})

type Mocks struct {
	In            *gbytes.Buffer
	Out           *gbytes.Buffer
	Err           *gbytes.Buffer
	CfService     *cfmysqlfakes.FakeCfService
	PortFinder    *cfmysqlfakes.FakePortFinder
	CliConnection *pluginfakes.FakeCliConnection
	MysqlRunner   *cfmysqlfakes.FakeMysqlRunner
}

func NewPluginAndMocks() (*MysqlPlugin, Mocks) {
	mocks := Mocks{
		In:            gbytes.NewBuffer(),
		Out:           gbytes.NewBuffer(),
		Err:           gbytes.NewBuffer(),
		CfService:     new(cfmysqlfakes.FakeCfService),
		CliConnection: new(pluginfakes.FakeCliConnection),
		MysqlRunner:   new(cfmysqlfakes.FakeMysqlRunner),
		PortFinder:    new(cfmysqlfakes.FakePortFinder),
	}

	mysqlPlugin := NewMysqlPlugin(PluginConf{
		mocks.In,
		mocks.Out,
		mocks.Err,
		mocks.CfService,
		mocks.MysqlRunner,
		mocks.PortFinder,
	})

	return mysqlPlugin, mocks
}
