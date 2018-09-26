package cfmysql_test

import (
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("CfService", func() {
	var apiClient *cfmysqlfakes.FakeApiClient
	var service CfService
	var cliConnection *pluginfakes.FakeCliConnection
	var sshRunner *cfmysqlfakes.FakeSshRunner
	var portWaiter *cfmysqlfakes.FakePortWaiter
	var mockHttp *cfmysqlfakes.FakeHttpWrapper
	var mockRand *cfmysqlfakes.FakeRandWrapper
	var appList []plugin_models.GetAppsModel
	var serviceKey models.ServiceKey
	var expectedMysqlService MysqlService
	var logWriter *gbytes.Buffer

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		cliConnection.GetCurrentSpaceReturns(plugin_models.Space{
			SpaceFields: plugin_models.SpaceFields{
				Guid: "space-guid-a",
				Name: "space is the place",
			},
		}, nil)

		apiClient = new(cfmysqlfakes.FakeApiClient)
		sshRunner = new(cfmysqlfakes.FakeSshRunner)
		portWaiter = new(cfmysqlfakes.FakePortWaiter)
		mockHttp = new(cfmysqlfakes.FakeHttpWrapper)
		mockRand = new(cfmysqlfakes.FakeRandWrapper)
		logWriter = gbytes.NewBuffer()

		service = NewCfService(apiClient, sshRunner, portWaiter, mockHttp, mockRand, logWriter)

		appList = []plugin_models.GetAppsModel{
			{
				Name: "app-name-1",
			},
			{
				Name: "app-name-2",
			},
		}

		serviceKey = models.ServiceKey{
			ServiceInstanceGuid: "service-instance-guid",
			Uri:                 "uri",
			DbName:              "db-name",
			Hostname:            "hostname",
			Port:                "2342",
			Username:            "username",
			Password:            "password",
			CaCert:              "ca-cert",
		}

		expectedMysqlService = MysqlService{
			Name:     "service-instance-name",
			Hostname: "hostname",
			Port:     "2342",
			DbName:   "db-name",
			Username: "username",
			Password: "password",
			CaCert:   "ca-cert",
		}
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
			Name:     "database-a",
			Hostname: "database-a.host",
			Port:     "3306",
			DbName:   "dbname-a",
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
				mockRand := new(cfmysqlfakes.FakeRandWrapper)
				logWriter := gbytes.NewBuffer()

				mockRand.IntnReturns(1)

				service := NewCfService(apiClient, sshRunner, portWaiter, mockHttp, mockRand, logWriter)

				sshRunner.OpenSshTunnelStub = notifyWhenGoroutineCalled

				service.OpenSshTunnel(cliConnection, mysqlService, appList, 4242)
				<-openSshTunnelCalled
				close(done)

				Expect(sshRunner.OpenSshTunnelCallCount()).To(Equal(1))

				calledCliConnection, calledService, calledAppName, calledPort := sshRunner.OpenSshTunnelArgsForCall(0)
				Expect(mockRand.IntnCallCount()).To(Equal(1))
				Expect(mockRand.IntnArgsForCall(0)).To(Equal(2))

				Expect(calledCliConnection).To(Equal(cliConnection))
				Expect(calledService).To(Equal(mysqlService))
				Expect(calledAppName).To(Equal("app-name-2"))
				Expect(calledPort).To(Equal(4242))
			}, 0.2)

			It("Blocks until the tunnel is open", func() {
				cliConnection.CliCommandWithoutTerminalOutputStub = nil
				service.OpenSshTunnel(cliConnection, mysqlService, appList, 4242)

				Expect(portWaiter.WaitUntilOpenCallCount()).To(Equal(1))
				Expect(portWaiter.WaitUntilOpenArgsForCall(0)).To(Equal(4242))
			})
		})
	})

	Context("GetService", func() {
		var instance models.ServiceInstance
		BeforeEach(func() {
			instance = models.ServiceInstance{
				Name:      "service-instance-name",
				Guid:      "service-instance-guid",
				SpaceGuid: "space-guid",
			}
		})

		Context("When the service instance is not found", func() {
			It("Returns an error", func() {
				apiClient.GetServiceReturns(models.ServiceInstance{}, errors.New("PC LOAD LETTER"))

				mysqlService, err := service.GetService(cliConnection, "service-name")

				Expect(mysqlService).To(Equal(MysqlService{}))
				Expect(err).To(Equal(errors.New("unable to retrieve metadata for service service-name: PC LOAD LETTER")))
				Expect(cliConnection.GetCurrentSpaceCallCount()).To(Equal(1))

				calledConnection, calledSpaceGuid, calledName := apiClient.GetServiceArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledSpaceGuid).To(Equal("space-guid-a"))
				Expect(calledName).To(Equal("service-name"))
			})
		})

		Context("When service and key are found", func() {
			It("Returns credentials", func() {
				apiClient.GetServiceReturns(instance, nil)
				apiClient.GetServiceKeyReturns(serviceKey, true, nil)

				mysqlService, err := service.GetService(cliConnection, "service-instance-name")

				Expect(err).To(BeNil())
				Expect(mysqlService).To(Equal(expectedMysqlService))

				calledConnection, calledSpaceGuid, calledName := apiClient.GetServiceArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledSpaceGuid).To(Equal("space-guid-a"))
				Expect(calledName).To(Equal("service-instance-name"))

				calledConnection, calledInstanceGuid, calledKeyName := apiClient.GetServiceKeyArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledInstanceGuid).To(Equal(instance.Guid))
				Expect(calledKeyName).To(Equal("cf-mysql"))
			})
		})

		Context("When the service key does not yet exist", func() {
			It("Creates the key and returns credentials", func() {
				apiClient.GetServiceReturns(instance, nil)
				apiClient.GetServiceKeyReturns(models.ServiceKey{}, false, nil)
				apiClient.CreateServiceKeyReturns(serviceKey, nil)

				mysqlService, err := service.GetService(cliConnection, "service-instance-name")

				Expect(err).To(BeNil())
				Expect(mysqlService).To(Equal(expectedMysqlService))

				calledConnection, calledSpaceGuid, calledName := apiClient.GetServiceArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledSpaceGuid).To(Equal("space-guid-a"))
				Expect(calledName).To(Equal("service-instance-name"))

				calledConnection, calledInstanceGuid, calledKeyName := apiClient.CreateServiceKeyArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledInstanceGuid).To(Equal(instance.Guid))
				Expect(calledKeyName).To(Equal("cf-mysql"))

				Expect(logWriter).To(gbytes.Say("Creating new service key cf-mysql for service-instance-name...\n"))
			})
		})

		Context("When the key cannot be created", func() {
			It("Returns an error", func() {
				apiClient.GetServiceReturns(instance, nil)
				apiClient.GetServiceKeyReturns(models.ServiceKey{}, false, nil)
				apiClient.CreateServiceKeyReturns(models.ServiceKey{}, errors.New("PC LOAD LETTER"))

				mysqlService, err := service.GetService(cliConnection, "service-instance-name")

				Expect(err).To(Equal(errors.New("unable to create service key: PC LOAD LETTER")))
				Expect(mysqlService).To(Equal(MysqlService{}))

				calledConnection, calledSpaceGuid, calledName := apiClient.GetServiceArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledSpaceGuid).To(Equal("space-guid-a"))
				Expect(calledName).To(Equal("service-instance-name"))

				calledConnection, calledInstanceGuid, calledKeyName := apiClient.CreateServiceKeyArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledInstanceGuid).To(Equal(instance.Guid))
				Expect(calledKeyName).To(Equal("cf-mysql"))
			})
		})

		Context("When the key cannot be retrieved", func() {
			It("Returns an error", func() {
				apiClient.GetServiceReturns(instance, nil)
				apiClient.GetServiceKeyReturns(models.ServiceKey{}, false, errors.New("PC LOAD LETTER"))

				mysqlService, err := service.GetService(cliConnection, "service-instance-name")

				Expect(err).To(Equal(errors.New("unable to retrieve service key: PC LOAD LETTER")))
				Expect(mysqlService).To(Equal(MysqlService{}))

				calledConnection, calledSpaceGuid, calledName := apiClient.GetServiceArgsForCall(0)
				Expect(calledConnection).To(Equal(cliConnection))
				Expect(calledSpaceGuid).To(Equal("space-guid-a"))
				Expect(calledName).To(Equal("service-instance-name"))
			})
		})

		Context("When the current space cannot be retrieved", func() {
			It("Returns an error", func() {
				cliConnection.GetCurrentSpaceReturns(plugin_models.Space{}, errors.New("PC LOAD LETTER"))

				mysqlService, err := service.GetService(cliConnection, "service-instance-name")

				Expect(err).To(Equal(errors.New("unable to retrieve current space: PC LOAD LETTER")))
				Expect(mysqlService).To(Equal(MysqlService{}))
				Expect(cliConnection.GetCurrentSpaceCallCount()).To(Equal(1))
			})
		})
	})
})
