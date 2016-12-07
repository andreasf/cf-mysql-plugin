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
	var err error
	var allInstances []pluginModels.ServiceInstance
	nextUrl := "/v2/service_instances"

	for nextUrl != "" {
		instanceResponse, err := self.getFromCfApi(nextUrl, cliConnection)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve service instances: %s", err)
		}

		var instances []pluginModels.ServiceInstance
		nextUrl, instances, err = deserializeInstances(instanceResponse)
		allInstances = append(allInstances, instances...)
	}

	return allInstances, err
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

func deserializeInstances(jsonResponse []byte) (string, []pluginModels.ServiceInstance, error) {
	paginatedResources := new(resources.PaginatedServiceInstanceResources)
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return "", nil, fmt.Errorf("Unable to deserialize service instances: %s", err)
	}

	return paginatedResources.NextUrl, paginatedResources.ToModel(), nil
}

func (self *ApiClientImpl) GetServiceBindings(cliConnection plugin.CliConnection) ([]pluginModels.ServiceBinding, error) {
	var allBindings []pluginModels.ServiceBinding
	nextUrl := "/v2/service_bindings"

	for nextUrl != "" {
		bindingsResp, err := self.getFromCfApi(nextUrl, cliConnection)
		if err != nil {
			return nil, fmt.Errorf("Unable to call service bindings endpoint: %s", err)
		}

		var bindings []pluginModels.ServiceBinding
		nextUrl, bindings, err = deserializeBindings(bindingsResp)
		if err != nil {
			return nil, fmt.Errorf("Unable to deserialize service bindings: %s", err)
		}

		allBindings = append(allBindings, bindings...)
	}

	return allBindings, nil
}

func deserializeBindings(bindingResponse []byte) (string, []pluginModels.ServiceBinding, error) {
	paginatedResources := new(resources.PaginatedServiceBindingResources)
	err := json.Unmarshal(bindingResponse, paginatedResources)

	if err != nil {
		return "", nil, fmt.Errorf("Unable to deserialize service bindings: %s", err)
	}

	bindings, err := paginatedResources.ToModel()
	return paginatedResources.NextUrl, bindings, err
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