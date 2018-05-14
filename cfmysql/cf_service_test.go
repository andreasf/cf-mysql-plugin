package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/models"
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

		service = NewCfService(apiClient, sshRunner, portWaiter, mockHttp, mockRand)

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
		}

		expectedMysqlService = MysqlService{
			Name:     "service-instance-name",
			Hostname: "hostname",
			Port:     "2342",
			DbName:   "db-name",
			Username: "username",
			Password: "password",
		}
	})

	Context("GetMysqlServices: retrieving available MySQL services", func() {
		It("Gets a list of service instances and bindings", func() {
			apiClient.GetServiceInstancesReturns(mockInstances(), nil)
			apiClient.GetServiceBindingsReturns(mockBindings(), nil)

			services, err := service.GetMysqlServices(cliConnection)

			Expect(apiClient.GetServiceBindingsCallCount()).To(Equal(1))
			Expect(apiClient.GetServiceBindingsArgsForCall(0)).To(Equal(cliConnection))

			Expect(err).To(BeNil())
			Expect(services).To(HaveLen(2))

			Expect(services[0]).To(Equal(MysqlService{
				Name:     "database-a",
				Hostname: "database-a.host",
				Port:     "3306",
				DbName:   "dbname-a",
				Username: "username-a",
				Password: "password-a",
			}))

			Expect(services[1]).To(Equal(MysqlService{
				Name:     "database-b",
				Hostname: "database-b.host",
				Port:     "3307",
				DbName:   "dbname-b",
				Username: "username-b",
				Password: "password-b",
			}))
		})

		Context("When the current space can't be determined", func() {
			It("Returns an error", func() {
				cliConnection.GetCurrentSpaceReturns(plugin_models.Space{}, errors.New("PC LOAD LETTER"))

				services, err := service.GetMysqlServices(cliConnection)

				Expect(services).To(BeNil())
				Expect(err).To(Equal(errors.New("Error retrieving current space: PC LOAD LETTER")))
			})
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

				mockRand.IntnReturns(1)

				service := NewCfService(apiClient, sshRunner, portWaiter, mockHttp, mockRand)

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
				Expect(mysqlService).To(Equal(MysqlService{
					Name:     "service-instance-name",
					Hostname: "hostname",
					Port:     "2342",
					DbName:   "db-name",
					Username: "username",
					Password: "password",
				}))

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

func mockInstances() []models.ServiceInstance {
	return []models.ServiceInstance{
		{
			Name:      "database-a",
			Guid:      "service-instance-guid-a",
			SpaceGuid: "space-guid-a",
		},
		{
			Name:      "database-b",
			Guid:      "service-instance-guid-b",
			SpaceGuid: "space-guid-a",
		},
		{
			Name:      "unbound-database-c",
			Guid:      "service-instance-guid-c",
			SpaceGuid: "space-guid-a",
		},
		{
			Name:      "redis-d",
			Guid:      "service-instance-guid-d",
			SpaceGuid: "space-guid-a",
		},
		{
			Name:      "database-e",
			Guid:      "service-instance-guid-e",
			SpaceGuid: "space-guid-b",
		},
	}
}

func mockBindings() []models.ServiceBinding {
	return []models.ServiceBinding{
		{
			ServiceInstanceGuid: "service-instance-guid-a",
			Uri:                 "mysql://username-a:password-a@database-a.host:3306/dbname-a?reconnect=true",
			DbName:              "dbname-a",
			Hostname:            "database-a.host",
			Port:                "3306",
			Username:            "username-a",
			Password:            "password-a",
		},
		{
			ServiceInstanceGuid: "service-instance-guid-b",
			Uri:                 "mysql://username-b:password-b@database-b.host:3307/dbname-b?reconnect=true",
			DbName:              "dbname-b",
			Hostname:            "database-b.host",
			Port:                "3307",
			Username:            "username-b",
			Password:            "password-b",
		},
		{
			ServiceInstanceGuid: "service-instance-guid-d",
			Uri:                 "",
			DbName:              "",
			Hostname:            "redis-host.name",
			Port:                "4321",
			Username:            "",
			Password:            "redis-password",
		},
		{
			ServiceInstanceGuid: "service-instance-guid-e",
			Uri:                 "mysql://username-e:password-e@database-e.host:54321/dbname-e?reconnect=true",
			DbName:              "dbname-e",
			Hostname:            "database-e.host",
			Port:                "54321",
			Username:            "username-e",
			Password:            "password-e",
		},
		{
			ServiceInstanceGuid: "service-instance-guid-f",
			Uri:                 "",
			DbName:              "",
			Hostname:            "",
			Port:                "",
			Username:            "",
			Password:            "",
		},
	}
}
