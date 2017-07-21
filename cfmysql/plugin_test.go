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
)

var _ = Describe("Plugin", func() {
	Context("When calling 'cf plugins'", func() {
		It("Shows the mysql plugin with the current version", func() {
			mysqlPlugin, _ := NewPluginAndMocks()

			Expect(mysqlPlugin.GetMetadata().Name).To(Equal("mysql"))
			Expect(mysqlPlugin.GetMetadata().Version).To(Equal(plugin.VersionType{
				Major: 1,
				Minor: 4,
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
		Context("With databases available", func() {
			var serviceA, serviceB MysqlService

			BeforeEach(func() {
				serviceA = MysqlService{
					Name:     "database-a",
					Hostname: "database-a.host",
					Port:     "123",
					DbName:   "dbname-a",
				}
				serviceB = MysqlService{
					Name:     "database-b",
					Hostname: "database-b.host",
					Port:     "234",
					DbName:   "dbname-b",
				}
			})

			It("Lists the available MySQL databases", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()
				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say("MySQL databases bound to an app:\n\ndatabase-a\ndatabase-b\n"))
				Expect(mocks.In).To(gbytes.Say(""))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(0))
			})
		})

		Context("With no databases available", func() {
			It("Tells the user that databases must be bound to a started app", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("No MySQL databases available. Please bind your database services to a started app to make them available to 'cf mysql'."))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(0))
			})
		})

		Context("With failing API calls", func() {
			It("Shows an error message", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns(nil, fmt.Errorf("foo"))

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("Unable to retrieve services: foo\n"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})
	})

	Context("When calling 'cf mysql db-name'", func() {
		var serviceA, serviceB MysqlService

		BeforeEach(func() {
			serviceA = MysqlService{
				Name:     "database-a",
				Hostname: "database-a.host",
				Port:     "123",
				DbName:   "dbname-a",
				Username: "username",
				Password: "password",
			}
			serviceB = MysqlService{
				Name:     "database-b",
				Hostname: "database-b.host",
				Port:     "234",
				DbName:   "dbname-b",
			}
		})

		Context("When the database is available", func() {
			var app plugin_models.GetAppsModel

			BeforeEach(func() {
				app = plugin_models.GetAppsModel{
					Name: "app-name",
				}
			})

			It("Opens an SSH tunnel through a started app", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app}, nil)
				mocks.PortFinder.GetPortReturns(2342)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
				Expect(mocks.CfService.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppName, localPort := mocks.CfService.OpenSshTunnelArgsForCall(0)
				Expect(calledCliConnection).To(Equal(mocks.CliConnection))
				Expect(calledService).To(Equal(serviceA))
				Expect(calledAppName).To(Equal("app-name"))
				Expect(localPort).To(Equal(2342))
			})

			It("Opens a MySQL client connecting through the tunnel", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app}, nil)
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

					mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
					mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app}, nil)
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

		Context("When the database is not available", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "db-name"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("^FAILED\nService 'db-name' is not bound to an app, not a MySQL database or does not exist in the current space.\n$"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})

		Context("When the GetMysqlServicesReturns returns an error", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns(nil, fmt.Errorf("PC LOAD LETTER"))

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "db-name"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("^FAILED\nUnable to retrieve services: PC LOAD LETTER\n$"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})

		Context("When there are no started apps", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say("^$"))
				Expect(mocks.Err).To(gbytes.Say("^FAILED\nUnable to connect to 'database-a': no started apps in current space\n$"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})

		Context("When the GetStartedApps returns an error", func() {
			It("Shows an error message and exits with 1", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
				mocks.CfService.GetStartedAppsReturns(nil, fmt.Errorf("PC LOAD LETTER"))

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysql", "database-a"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
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
		Context("With databases available", func() {
			var serviceA, serviceB MysqlService

			BeforeEach(func() {
				serviceA = MysqlService{
					Name:     "database-a",
					Hostname: "database-a.host",
					Port:     "123",
					DbName:   "dbname-a",
				}
				serviceB = MysqlService{
					Name:     "database-b",
					Hostname: "database-b.host",
					Port:     "234",
					DbName:   "dbname-b",
				}
			})

			It("Lists the available MySQL databases", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say("MySQL databases bound to an app:\n\ndatabase-a\ndatabase-b\n"))
				Expect(mocks.Err).To(gbytes.Say(""))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(0))
			})
		})

		Context("With no databases available", func() {
			It("Tells the user that databases must be bound to a started app", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{}, nil)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("No MySQL databases available. Please bind your database services to a started app to make them available to 'cf mysqldump'."))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(0))
			})
		})

		Context("With failing API calls", func() {
			It("Shows an error message", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns(nil, fmt.Errorf("foo"))

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.Out).To(gbytes.Say(""))
				Expect(mocks.Err).To(gbytes.Say("Unable to retrieve services: foo\n"))
				Expect(mysqlPlugin.GetExitCode()).To(Equal(1))
			})
		})
	})

	Context("When calling 'cf mysql db-name'", func() {
		var serviceA, serviceB MysqlService

		BeforeEach(func() {
			serviceA = MysqlService{
				Name:     "database-a",
				Hostname: "database-a.host",
				Port:     "123",
				DbName:   "dbname-a",
				Username: "username",
				Password: "password",
			}
			serviceB = MysqlService{
				Name:     "database-b",
				Hostname: "database-b.host",
				Port:     "234",
				DbName:   "dbname-b",
			}
		})

		Context("When the database is available", func() {
			var app plugin_models.GetAppsModel

			BeforeEach(func() {
				app = plugin_models.GetAppsModel{
					Name: "app-name",
				}
			})

			It("Opens an SSH tunnel through a started app", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app}, nil)
				mocks.PortFinder.GetPortReturns(2342)

				mysqlPlugin.Run(mocks.CliConnection, []string{"mysqldump", "database-a"})

				Expect(mocks.CfService.GetMysqlServicesCallCount()).To(Equal(1))
				Expect(mocks.CfService.GetStartedAppsCallCount()).To(Equal(1))
				Expect(mocks.PortFinder.GetPortCallCount()).To(Equal(1))
				Expect(mocks.CfService.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppName, localPort := mocks.CfService.OpenSshTunnelArgsForCall(0)
				Expect(calledCliConnection).To(Equal(mocks.CliConnection))
				Expect(calledService).To(Equal(serviceA))
				Expect(calledAppName).To(Equal("app-name"))
				Expect(localPort).To(Equal(2342))
			})

			It("Opens mysqldump connecting through the tunnel", func() {
				mysqlPlugin, mocks := NewPluginAndMocks()

				mocks.CfService.GetMysqlServicesReturns([]MysqlService{serviceA, serviceB}, nil)
				mocks.CfService.GetStartedAppsReturns([]plugin_models.GetAppsModel{app}, nil)
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

	Context("When the plugin is being uninstalled", func() {
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
