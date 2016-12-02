package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	. "code.cloudfoundry.org/cli/plugin/models"
	"fmt"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"encoding/json"
)

//go:generate counterfeiter . ApiClient
type ApiClient interface {
	GetServiceInstances(cliConnection plugin.CliConnection) (*resources.PaginatedServiceInstanceResources, error)
	GetStartedApps(cliConnection plugin.CliConnection) ([]GetAppsModel, error)
}

type ApiClientImpl struct {
	HttpClient Http
}

func (self *ApiClientImpl) GetServiceInstances(cliConnection plugin.CliConnection) (*resources.PaginatedServiceInstanceResources, error) {
	instanceResponse, err := self.getFromCfApi("/v2/service_instances", cliConnection)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve service instances: %s", err)
	}

	return deserializeInstances(instanceResponse)
}

func (self *ApiClientImpl) getFromCfApi(path string, cliConnection plugin.CliConnection) ([]byte, error) {
	endpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		return nil, fmt.Errorf("Unable to get API endpoint: %s", err)
	}

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return nil, fmt.Errorf("Unable to get access token: %s", err)
	}

	return self.HttpClient.Get(endpoint + path, accessToken)
}

func deserializeInstances(jsonResponse []byte) (*resources.PaginatedServiceInstanceResources, error) {
	paginatedResources := new(resources.PaginatedServiceInstanceResources)
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service instances: %s", err)
	}

	return paginatedResources, nil
}

func (self *ApiClientImpl) GetStartedApps(cliConnection plugin.CliConnection) ([]GetAppsModel, error) {
	apps, err := cliConnection.GetApps()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve apps: %s", err)
	}

	startedApps := make([]GetAppsModel, 0, len(apps))

	for _, app := range (apps) {
		if app.State == "started" {
			startedApps = append(startedApps, app)
		}
	}

	return startedApps, nil
}

func NewApiClient() *ApiClientImpl {
	return &ApiClientImpl{
		HttpClient: new(HttpWrapper),
	}
}