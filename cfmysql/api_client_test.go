package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
)

var _ = Describe("ApiClient", func() {
	var cliConnection *pluginfakes.FakeCliConnection
	var apiClient *ApiClientImpl

	BeforeEach(func() {
		cliConnection = new(pluginfakes.FakeCliConnection)
		apiClient = &ApiClientImpl{}
	})

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
