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
	var apiClient *SdkApiClient
	var cliConnection *pluginfakes.FakeCliConnection
	var sshRunner *cfmysqlfakes.FakeSshRunner
	var portWaiter *cfmysqlfakes.FakePortWaiter

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		cliConnection.CliCommandWithoutTerminalOutputStub = mockCfCurl
		sshRunner = new(cfmysqlfakes.FakeSshRunner)
		portWaiter = new(cfmysqlfakes.FakePortWaiter)
		apiClient = &SdkApiClient{
			SshRunner: sshRunner,
			PortWaiter: portWaiter,
		}
	})

	Context("GetMysqlServices: retrieving available MySQL services", func() {

		It("Gets a list of bindings", func() {
			paginatedResources, err := apiClient.GetServiceBindings(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(3))
		})

		It("Gets a list of instances", func() {
			paginatedResources, err := apiClient.GetServiceInstances(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(4))
		})

		It("Gets a list of service instances and bindings", func() {
			services, err := apiClient.GetMysqlServices(cliConnection)

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

			Expect(cliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(2))
			Expect(cliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)).To(Equal([]string{"curl", "/v2/service_bindings"}))
			Expect(cliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)).To(Equal([]string{"curl", "/v2/service_instances"}))
		})
	})

	Context("GetStartedApps", func() {
		Context("When the API returns the list of apps", func() {
			app1 := plugin_models.GetAppsModel{
				Name: "app-name-a",
				State: "stopped",
			}
			app2 := plugin_models.GetAppsModel{
				Name: "app-name-b",
				State: "started",
			}
			app3 := plugin_models.GetAppsModel{
				Name: "app-name-c",
				State: "started",
			}

			It("Returns the list of started apps", func() {
				cliConnection.GetAppsReturns([]plugin_models.GetAppsModel{app1, app2, app3}, nil)

				startedApps, err := apiClient.GetStartedApps(cliConnection)

				Expect(err).To(BeNil())
				Expect(startedApps).To(HaveLen(2))
				Expect(startedApps[0].Name).To(Equal("app-name-b"))
				Expect(startedApps[1].Name).To(Equal("app-name-c"))
			})
		})

		Context("When the API returns an error", func() {
			It("Returns an error", func() {
				cliConnection.GetAppsReturns(nil, errors.New("PC LOAD LETTER"))

				_, err := apiClient.GetStartedApps(cliConnection)

				Expect(err).To(Equal(errors.New("Unable to retrieve apps: PC LOAD LETTER")))
			})
		})
	})

	Context("OpenSshTunnel", func() {
		service := MysqlService{
			Name: "database-a",
			AppName: "",
			Hostname: "database-a.host",
			Port: "3306",
			DbName: "dbname-a",
			Username: "username-a",
			Password: "password-a",
		}

		Context("When opening the tunnel", func() {
			notifyWhenGoroutineCalled := func(cliConnection plugin.CliConnection, toService MysqlService, throughApp string, localPort int, doneChan chan bool) {
				doneChan <- true
			}

			It("Runs the SSH runner in a goroutine", func(done Done) {
				cliConnection.CliCommandWithoutTerminalOutputStub = nil
				sshRunner.OpenSshTunnelStub = notifyWhenGoroutineCalled

				apiClient.OpenSshTunnel(cliConnection, service, "app-name", 4242)
				<-apiClient.GetTunnelChan()
				close(done)

				Expect(sshRunner.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppName, calledPort, doneChan := sshRunner.OpenSshTunnelArgsForCall(0)
				Expect(calledCliConnection).To(Equal(cliConnection))
				Expect(calledService).To(Equal(service))
				Expect(calledAppName).To(Equal("app-name"))
				Expect(calledPort).To(Equal(4242))
				Expect(doneChan).To(Equal(apiClient.GetTunnelChan()))
			}, 0.2)

			It("Blocks until the tunnel is open", func() {
				cliConnection.CliCommandWithoutTerminalOutputStub = nil
				apiClient.OpenSshTunnel(cliConnection, service, "app-name", 4242)

				Expect(portWaiter.WaitUntilOpenCallCount()).To(Equal(1))
				Expect(portWaiter.WaitUntilOpenArgsForCall(0)).To(Equal(4242))
			})
		})
	})
})

func mockCfCurl(args... string) ([]string, error) {
	if SliceEquals(args, []string{"curl", "/v2/service_bindings"}) {
		return test_resources.LoadResourceLines("test_resources/service_bindings.json"), nil
	}
	if SliceEquals(args, []string{"curl", "/v2/service_instances"}) {
		return test_resources.LoadResourceLines("test_resources/service_instances.json"), nil
	}

	return nil, fmt.Errorf("No mock defined for %s", args)
}

func SliceEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
