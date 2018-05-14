package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"

	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/test_resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/models"
	"io"
	"bytes"
)

var _ = Describe("ApiClient", func() {
	var cliConnection *pluginfakes.FakeCliConnection
	var apiClient ApiClient
	var mockHttp *cfmysqlfakes.FakeHttpWrapper

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		mockHttp = new(cfmysqlfakes.FakeHttpWrapper)
		mockHttp.GetStub = func(url string, accessToken string, skipSsl bool) ([]byte, error) {
			switch url {
			case "https://cf.api.url/v2/service_instances/service-instance-guid/service_keys?q=name%3Aservice-key-name":
				return test_resources.LoadResource("test_resources/service_key.json"), nil
			case "https://cf.api.url/v2/service_instances/service-instance-guid/service_keys?q=name%3Ano-such-key":
				return test_resources.LoadResource("test_resources/service_key_empty.json"), nil
			case "https://cf.api.url/v2/spaces/space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Aservice-name-a":
				return test_resources.LoadResource("test_resources/service_instance.json"), nil
			case "https://cf.api.url/v2/spaces/space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Ano-such-service":
				return test_resources.LoadResource("test_resources/service_instance_empty.json"), nil
			default:
				return nil, fmt.Errorf("URL not handled in mock: %s", url)
			}
		}

		mockHttp.PostStub = func(url string, body io.Reader, accessToken string, skipSsl bool) ([]byte, error) {
			switch url {
			case "https://cf.api.url/v2/service_keys":
				return test_resources.LoadResource("test_resources/service_key_created.json"), nil
			default:
				return nil, fmt.Errorf("URL not handled in mock: %s", url)
			}
		}

		apiClient = NewApiClient(mockHttp)

		cliConnection.ApiEndpointReturns("https://cf.api.url", nil)
		cliConnection.AccessTokenReturns("bearer my-secret-token", nil)
		cliConnection.IsSSLDisabledReturns(true, nil)
	})

	Describe("GetStartedApps", func() {
		Context("When the API returns the list of apps", func() {
			app1 := plugin_models.GetAppsModel{
				Name:  "app-name-a",
				State: "stopped",
			}
			app2 := plugin_models.GetAppsModel{
				Name:  "app-name-b",
				State: "started",
			}
			app3 := plugin_models.GetAppsModel{
				Name:  "app-name-c",
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

	Describe("GetService", func() {
		Context("When the API returns a matching service", func() {
			It("Returns service info", func() {
				instance, err := apiClient.GetService(cliConnection, "space-guid", "service-name-a")

				Expect(err).To(BeNil())
				Expect(instance).To(Equal(models.ServiceInstance{
					Name:      "service-name-a",
					Guid:      "service-instance-guid-a",
					SpaceGuid: "space-guid",
				}))
			})
		})

		Context("When no matching service instance is returned", func() {
			It("Returns an error", func() {
				instance, err := apiClient.GetService(cliConnection, "space-guid", "no-such-service")

				Expect(err).To(Equal(errors.New("no-such-service not found in current space")))
				Expect(instance).To(Equal(models.ServiceInstance{}))
			})
		})
	})

	Describe("GetServiceKey", func() {
		Context("When the API returns a key", func() {
			It("Returns the key", func() {
				serviceKey, found, err := apiClient.GetServiceKey(cliConnection, "service-instance-guid", "service-key-name")

				Expect(found).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(serviceKey).To(Equal(models.ServiceKey{
					ServiceInstanceGuid: "service-instance-guid",
					Uri:                 "uri",
					DbName:              "db-name",
					Hostname:            "hostname",
					Port:                "3306",
					Username:            "username",
					Password:            "password",
				}))
			})
		})

		Context("When no key was found for the given service guid and key name", func() {
			It("Returns not found", func() {
				serviceKey, found, err := apiClient.GetServiceKey(cliConnection, "service-instance-guid", "no-such-key")

				Expect(found).To(BeFalse())
				Expect(err).To(BeNil())
				Expect(serviceKey).To(Equal(models.ServiceKey{}))
			})
		})
	})

	Describe("CreateServiceKey", func() {
		Context("When the API returns a key", func() {
			It("Returns the key", func() {
				serviceKey, err := apiClient.CreateServiceKey(cliConnection, "service-instance-guid", "service-key-name")

				url, body, accessToken, sslDisabled := mockHttp.PostArgsForCall(0)
				Expect(url).To(Equal("https://cf.api.url/v2/service_keys"))
				Expect(body).To(Equal(bytes.NewBuffer([]byte("{\"name\":\"service-key-name\",\"service_instance_guid\":\"service-instance-guid\"}"))))
				Expect(accessToken).To(Equal("bearer my-secret-token"))
				Expect(sslDisabled).To(BeTrue())

				Expect(err).To(BeNil())
				Expect(serviceKey).To(Equal(models.ServiceKey{
					ServiceInstanceGuid: "service-instance-guid",
					Uri:                 "uri",
					DbName:              "db-name",
					Hostname:            "hostname",
					Port:                "3306",
					Username:            "username",
					Password:            "password",
				}))
			})
		})
	})
})
