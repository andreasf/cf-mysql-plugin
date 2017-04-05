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
)

var _ = Describe("ApiClient", func() {
	var cliConnection *pluginfakes.FakeCliConnection
	var apiClient *ApiClientImpl
	var mockHttp *cfmysqlfakes.FakeHttp

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		mockHttp = new(cfmysqlfakes.FakeHttp)
		mockHttp.GetStub = func(url string, accessToken string, skipSsl bool) ([]byte, error) {
			switch url {
			case "https://cf.api.url/v2/service_bindings":
				return test_resources.LoadResource("test_resources/service_bindings.json"), nil
			case "https://cf.api.url/v2/service_bindings?page=2":
				return test_resources.LoadResource("test_resources/service_bindings_page2.json"), nil
			case "https://cf.api.url/v2/service_instances":
				return test_resources.LoadResource("test_resources/service_instances.json"), nil
			case "https://cf.api.url/v2/service_instances?page=2":
				return test_resources.LoadResource("test_resources/service_instances_page2.json"), nil
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

	Describe("GetServiceInstances", func() {
		It("Gets a list of instances", func() {
			cliConnection.ApiEndpointReturns("https://cf.api.url", nil)
			cliConnection.AccessTokenReturns("bearer my-secret-token", nil)

			instances, err := apiClient.GetServiceInstances(cliConnection)

			Expect(err).To(BeNil())
			Expect(instances).To(HaveLen(5))

			Expect(cliConnection.AccessTokenCallCount()).To(Equal(2))
			Expect(cliConnection.ApiEndpointCallCount()).To(Equal(2))

			Expect(mockHttp.GetCallCount()).To(Equal(2))
			url, access_token := mockHttp.GetArgsForCall(0)
			Expect(url).To(Equal("https://cf.api.url/v2/service_instances"))
			Expect(access_token).To(Equal("bearer my-secret-token"))

			url2, access_token2 := mockHttp.GetArgsForCall(1)
			Expect(url2).To(Equal("https://cf.api.url/v2/service_instances?page=2"))
			Expect(access_token2).To(Equal("bearer my-secret-token"))
		})
	})

	Describe("GetServiceBindings", func() {
		Context("When the API returns a list of bindings", func() {
			It("Returns the list of bindings", func() {
				cliConnection.ApiEndpointReturns("https://cf.api.url", nil)
				cliConnection.AccessTokenReturns("bearer my-secret-token", nil)

				bindings, err := apiClient.GetServiceBindings(cliConnection)

				Expect(err).To(BeNil())
				Expect(bindings).To(HaveLen(6))
				Expect(bindings[0].Port).To(Equal("3306"))
				Expect(bindings[3].Port).To(Equal("54321"))

				Expect(cliConnection.AccessTokenCallCount()).To(Equal(2))
				Expect(cliConnection.ApiEndpointCallCount()).To(Equal(2))

				Expect(mockHttp.GetCallCount()).To(Equal(2))
				url, access_token := mockHttp.GetArgsForCall(0)
				Expect(url).To(Equal("https://cf.api.url/v2/service_bindings"))
				Expect(access_token).To(Equal("bearer my-secret-token"))

				url2, access_token2 := mockHttp.GetArgsForCall(1)
				Expect(url2).To(Equal("https://cf.api.url/v2/service_bindings?page=2"))
				Expect(access_token2).To(Equal("bearer my-secret-token"))
			})
		})
	})
})
