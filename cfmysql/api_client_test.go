package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/test_resources"
	"fmt"
)

var _ = Describe("ApiClient", func() {
	var cliConnection *pluginfakes.FakeCliConnection
	var apiClient *ApiClientImpl
	var mockHttp *cfmysqlfakes.FakeHttp

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		mockHttp = new(cfmysqlfakes.FakeHttp)
		mockHttp.GetStub = func(url string, accessToken string) ([]byte, error) {
			switch url {
			case "https://cf.api.url/v2/service_instances":
				return test_resources.LoadResource("test_resources/service_instances.json"), nil
			default:
				return nil, fmt.Errorf("URL not handled in mock: %s", url)
			}
		}

		apiClient = &ApiClientImpl{
			HttpClient: mockHttp,
		}
	})

	Describe("GetStartedApps", func() {
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

	Describe("GetServiceInstances", func() {
		It("Gets a list of instances", func() {
			cliConnection.ApiEndpointReturns("https://cf.api.url", nil)
			cliConnection.AccessTokenReturns("bearer my-secret-token", nil)

			paginatedResources, err := apiClient.GetServiceInstances(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(4))

			Expect(cliConnection.AccessTokenCallCount()).To(Equal(1))
			Expect(cliConnection.ApiEndpointCallCount()).To(Equal(1))

			Expect(mockHttp.GetCallCount()).To(Equal(1))
			url, access_token := mockHttp.GetArgsForCall(0)
			Expect(url).To(Equal("https://cf.api.url/v2/service_instances"))
			Expect(access_token).To(Equal("bearer my-secret-token"))
		})
	})
})
