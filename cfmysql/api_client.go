package cfmysql

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
	sdkModels "code.cloudfoundry.org/cli/plugin/models"
	pluginModels "github.com/andreasf/cf-mysql-plugin/cfmysql/models"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/resources"
	"net/url"
	"io"
	"bytes"
)

//go:generate counterfeiter . ApiClient
type ApiClient interface {
	GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error)
	GetService(cliConnection plugin.CliConnection, spaceGuid string, name string) (pluginModels.ServiceInstance, error)
	GetServiceKey(cliConnection plugin.CliConnection, serviceInstanceGuid string, keyName string) (key pluginModels.ServiceKey, found bool, err error)
	CreateServiceKey(cliConnection plugin.CliConnection, serviceInstanceGuid string, keyName string) (pluginModels.ServiceKey, error)
}

func NewApiClient(httpClient HttpWrapper) *apiClient {
	return &apiClient{
		httpClient: httpClient,
	}
}

type apiClient struct {
	httpClient HttpWrapper
}

func (self *apiClient) GetService(cliConnection plugin.CliConnection, spaceGuid string, name string) (pluginModels.ServiceInstance, error) {
	path := fmt.Sprintf(
		"/v2/spaces/%s/service_instances?return_user_provided_service_instances=true&q=name%%3A%s",
		spaceGuid,
		url.QueryEscape(name),
	)

	instanceResponse, err := self.getFromCfApi(path, cliConnection)
	if err != nil {
		return pluginModels.ServiceInstance{}, fmt.Errorf("error retrieving service instance: %s", err)
	}

	_, instances, err := deserializeInstances(instanceResponse)
	if err != nil {
		return pluginModels.ServiceInstance{}, fmt.Errorf("error deserializing service instances: %s", err)
	}

	if len(instances) == 0 {
		return pluginModels.ServiceInstance{}, fmt.Errorf("%s not found in current space", name)
	}

	return instances[0], nil
}

func (self *apiClient) GetServiceKey(cliConnection plugin.CliConnection, serviceInstanceGuid string, keyName string) (pluginModels.ServiceKey, bool, error) {
	path := fmt.Sprintf(
		"/v2/service_instances/%s/service_keys?q=name%%3A%s",
		serviceInstanceGuid,
		url.QueryEscape(keyName),
	)

	keyResponse, err := self.getFromCfApi(path, cliConnection)
	if err != nil {
		return pluginModels.ServiceKey{}, false, fmt.Errorf("error retrieving service key: %s", err)
	}

	serviceKeys, err := deserializeServiceKeys(keyResponse)
	if err != nil {
		return pluginModels.ServiceKey{}, false, fmt.Errorf("error deserializing service key response: %s", err)
	}

	if len(serviceKeys) == 0 {
		return pluginModels.ServiceKey{}, false, nil
	}

	return serviceKeys[0], true, nil
}

func (self *apiClient) CreateServiceKey(cliConnection plugin.CliConnection, serviceInstanceGuid string, keyName string) (pluginModels.ServiceKey, error) {
	content := ServiceKeyRequest{
		Name:                keyName,
		ServiceInstanceGuid: serviceInstanceGuid,
	}

	body, err := json.Marshal(content)
	if err != nil {
		return pluginModels.ServiceKey{}, fmt.Errorf("error serializing request body: %s", err)
	}

	response, err := self.postToCfApi("/v2/service_keys", bytes.NewBuffer(body), cliConnection)
	if err != nil {
		return pluginModels.ServiceKey{}, fmt.Errorf("error creating service key: %s", err)
	}

	serviceKey, err := deserializeServiceKey(response)
	if err != nil {
		return pluginModels.ServiceKey{}, fmt.Errorf("error deserializing service key response: %s", err)
	}

	return serviceKey, nil
}

func (self *apiClient) GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error) {
	apps, err := cliConnection.GetApps()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve apps: %s", err)
	}

	startedApps := make([]sdkModels.GetAppsModel, 0, len(apps))

	for _, app := range apps {
		if app.State == "started" {
			startedApps = append(startedApps, app)
		}
	}

	return startedApps, nil
}

func (self *apiClient) getFromCfApi(path string, cliConnection plugin.CliConnection) ([]byte, error) {
	endpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		return nil, fmt.Errorf("Unable to get API endpoint: %s", err)
	}

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return nil, fmt.Errorf("Unable to get access token: %s", err)
	}

	sslDisabled, err := cliConnection.IsSSLDisabled()
	if err != nil {
		return nil, fmt.Errorf("Unable to check SSL status: %s", err)
	}

	return self.httpClient.Get(endpoint+path, accessToken, sslDisabled)
}

func (self *apiClient) postToCfApi(path string, body io.Reader, cliConnection plugin.CliConnection) ([]byte, error) {
	endpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		return nil, fmt.Errorf("Unable to get API endpoint: %s", err)
	}

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return nil, fmt.Errorf("Unable to get access token: %s", err)
	}

	sslDisabled, err := cliConnection.IsSSLDisabled()
	if err != nil {
		return nil, fmt.Errorf("Unable to check SSL status: %s", err)
	}

	return self.httpClient.Post(endpoint+path, body, accessToken, sslDisabled)
}

func deserializeInstances(jsonResponse []byte) (string, []pluginModels.ServiceInstance, error) {
	paginatedResources := new(resources.PaginatedServiceInstanceResources)
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return "", nil, fmt.Errorf("Unable to deserialize service instances: %s", err)
	}

	return paginatedResources.NextUrl, paginatedResources.ToModel(), nil
}

func deserializeServiceKeys(keyResponse []byte) ([]pluginModels.ServiceKey, error) {
	paginatedResources := new(resources.PaginatedServiceKeyResources)
	err := json.Unmarshal(keyResponse, paginatedResources)
	if err != nil {
		return nil, fmt.Errorf("error deserializing service key response: %s", err)
	}

	serviceKeys, err := paginatedResources.ToModel()
	if err != nil {
		return nil, fmt.Errorf("error converting service key response: %s", err)
	}

	return serviceKeys, nil
}

func deserializeServiceKey(keyResponse []byte) (pluginModels.ServiceKey, error) {
	resource := new(resources.ServiceKeyResource)
	err := json.Unmarshal(keyResponse, resource)
	if err != nil {
		return pluginModels.ServiceKey{}, fmt.Errorf("error deserializing service key response: %s", err)
	}

	serviceKey, err := resource.ToModel()
	if err != nil {
		return pluginModels.ServiceKey{}, fmt.Errorf("error converting service key response: %s", err)
	}

	return serviceKey, nil
}

type ServiceKeyRequest struct {
	Name                string `json:"name"`
	ServiceInstanceGuid string `json:"service_instance_guid"`
}
