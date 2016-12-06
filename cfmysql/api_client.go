package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
	"encoding/json"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/resources"
	sdkModels "code.cloudfoundry.org/cli/plugin/models"
	pluginModels "github.com/andreasf/cf-mysql-plugin/cfmysql/models"
)

//go:generate counterfeiter . ApiClient
type ApiClient interface {
	GetServiceBindings(cliConnection plugin.CliConnection) ([]pluginModels.ServiceBinding, error)
	GetServiceInstances(cliConnection plugin.CliConnection) ([]pluginModels.ServiceInstance, error)
	GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error)
}

type ApiClientImpl struct {
	HttpClient Http
}

func (self *ApiClientImpl) GetServiceInstances(cliConnection plugin.CliConnection) ([]pluginModels.ServiceInstance, error) {
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

func deserializeInstances(jsonResponse []byte) ([]pluginModels.ServiceInstance, error) {
	paginatedResources := new(resources.PaginatedServiceInstanceResources)
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service instances: %s", err)
	}

	return paginatedResources.ToModel(), nil
}

func (self *ApiClientImpl) GetServiceBindings(cliConnection plugin.CliConnection) ([]pluginModels.ServiceBinding, error) {
	bindingsResp, err := self.getFromCfApi("/v2/service_bindings", cliConnection)
	if err != nil {
		return nil, fmt.Errorf("Unable to call service bindings endpoint: %s", err)
	}

	return deserializeBindings(bindingsResp)
}

func deserializeBindings(bindingResponse []byte) ([]pluginModels.ServiceBinding, error) {
	paginatedResources := new(resources.PaginatedServiceBindingResources)
	err := json.Unmarshal(bindingResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service bindings: %s", err)
	}

	return paginatedResources.ToModel()
}

func (self *ApiClientImpl) GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error) {
	apps, err := cliConnection.GetApps()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve apps: %s", err)
	}

	startedApps := make([]sdkModels.GetAppsModel, 0, len(apps))

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