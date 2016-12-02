package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/test_resources"
	"fmt"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"code.cloudfoundry.org/cli/plugin"
)

var _ = Describe("CfSdkClient", func() {
	var apiClient *cfmysqlfakes.FakeApiClient
	var service *CfServiceImpl
	var cliConnection *pluginfakes.FakeCliConnection
	var sshRunner *cfmysqlfakes.FakeSshRunner
	var portWaiter *cfmysqlfakes.FakePortWaiter
	var mockHttp *cfmysqlfakes.FakeHttp

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		cliConnection.ApiEndpointReturns("https://cf.api.url", nil)
		cliConnection.AccessTokenReturns("bearer my-secret-token", nil)

		apiClient = new(cfmysqlfakes.FakeApiClient)
		sshRunner = new(cfmysqlfakes.FakeSshRunner)
		portWaiter = new(cfmysqlfakes.FakePortWaiter)

		mockHttp = new(cfmysqlfakes.FakeHttp)
		mockHttp.GetStub = func(url string, accessToken string) ([]byte, error) {
			switch url {
			case "https://cf.api.url/v2/service_bindings":
				return test_resources.LoadResource("test_resources/service_bindings.json"), nil
			case "https://cf.api.url/v2/service_instances":
				return test_resources.LoadResource("test_resources/service_instances.json"), nil
			default:
				return nil, fmt.Errorf("URL not handled in mock: %s", url)
			}
		}

		service = &CfServiceImpl{
			ApiClient: apiClient,
			SshRunner: sshRunner,
			PortWaiter: portWaiter,
			HttpClient: mockHttp,
		}
	})

	Context("GetMysqlServices: retrieving available MySQL services", func() {
		It("Gets a list of bindings", func() {
			paginatedResources, err := service.GetServiceBindings(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(5))
			Expect(paginatedResources.Resources[0].Entity.Credentials.Port).To(Equal("3306"))
			Expect(paginatedResources.Resources[3].Entity.Credentials.Port).To(Equal("54321"))

			Expect(cliConnection.AccessTokenCallCount()).To(Equal(1))
			Expect(cliConnection.ApiEndpointCallCount()).To(Equal(1))

			Expect(mockHttp.GetCallCount()).To(Equal(1))
			url, access_token := mockHttp.GetArgsForCall(0)
			Expect(url).To(Equal("https://cf.api.url/v2/service_bindings"))
			Expect(access_token).To(Equal("bearer my-secret-token"))

		})

		It("Gets a list of instances", func() {
			cliConnection.ApiEndpointReturns("https://cf.api.url", nil)
			cliConnection.AccessTokenReturns("bearer my-secret-token", nil)

			paginatedResources, err := service.GetServiceInstances(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(4))

			Expect(cliConnection.AccessTokenCallCount()).To(Equal(1))
			Expect(cliConnection.ApiEndpointCallCount()).To(Equal(1))

			Expect(mockHttp.GetCallCount()).To(Equal(1))
			url, access_token := mockHttp.GetArgsForCall(0)
			Expect(url).To(Equal("https://cf.api.url/v2/service_instances"))
			Expect(access_token).To(Equal("bearer my-secret-token"))
		})

		It("Gets a list of service instances and bindings", func() {
			services, err := service.GetMysqlServices(cliConnection)

			Expect(err).To(BeNil())
			Expect(services).To(HaveLen(2))

			Expect(services[0]).To(Equal(MysqlService{
				Name: "database-a",
				AppName: "",
				Hostname: "database-a.host",
				Port: "3306",
				DbName: "dbname-a",
				Username: "username-a",
				Password: "password-a",
			}))

			Expect(services[1]).To(Equal(MysqlService{
				Name: "database-b",
				AppName: "",
				Hostname: "database-b.host",
				Port: "3307",
				DbName: "dbname-b",
				Username: "username-b",
				Password: "password-b",
			}))
		})
	})

	Context("GetStartedApps", func() {
		It("delegates the call to ApiClient", func() {
			expectedApps := []plugin_models.GetAppsModel{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			}
			expectedErr := errors.New("baz")

			apiClient.GetStartedAppsReturns(expectedApps, expectedErr)

			startedApps, err := service.GetStartedApps(cliConnection)

			Expect(startedApps).To(Equal(expectedApps))
			Expect(err).To(Equal(expectedErr))

			Expect(apiClient.GetStartedAppsCallCount()).To(Equal(1))
			Expect(apiClient.GetStartedAppsArgsForCall(0)).To(Equal(cliConnection))
		})
	})

	Context("OpenSshTunnel", func() {
		mysqlService := MysqlService{
			Name: "database-a",
			AppName: "",
			Hostname: "database-a.host",
			Port: "3306",
			DbName: "dbname-a",
			Username: "username-a",
			Password: "password-a",
		}
		openSshTunnelCalled := make(chan bool, 0)

		Context("When opening the tunnel", func() {
			notifyWhenGoroutineCalled := func(cliConnection plugin.CliConnection, toService MysqlService, throughApp string, localPort int) {
				openSshTunnelCalled <- true
			}

			It("Runs the SSH runner in a goroutine", func(done Done) {
				cliConnection := new(pluginfakes.FakeCliConnection)
				sshRunner := new(cfmysqlfakes.FakeSshRunner)
				portWaiter := new(cfmysqlfakes.FakePortWaiter)
				service := &CfServiceImpl{
					SshRunner: sshRunner,
					PortWaiter: portWaiter,
				}

				sshRunner.OpenSshTunnelStub = notifyWhenGoroutineCalled

				service.OpenSshTunnel(cliConnection, mysqlService, "app-name", 4242)
				<-openSshTunnelCalled
				close(done)

				Expect(sshRunner.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppName, calledPort := sshRunner.OpenSshTunnelArgsForCall(0)
				Expect(calledCliConnection).To(Equal(cliConnection))
				Expect(calledService).To(Equal(mysqlService))
				Expect(calledAppName).To(Equal("app-name"))
				Expect(calledPort).To(Equal(4242))
			}, 0.2)

			It("Blocks until the tunnel is open", func() {
				cliConnection.CliCommandWithoutTerminalOutputStub = nil
				service.OpenSshTunnel(cliConnection, mysqlService, "app-name", 4242)

				Expect(portWaiter.WaitUntilOpenCallCount()).To(Equal(1))
				Expect(portWaiter.WaitUntilOpenArgsForCall(0)).To(Equal(4242))
			})
		})
	})
})
