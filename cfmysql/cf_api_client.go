package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
	"strings"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"encoding/json"
	pluginResources "github.com/andreasf/cf-mysql-plugin/cfmysql/resources"
	. "code.cloudfoundry.org/cli/plugin/models"
)

//go:generate counterfeiter . ApiClient
type ApiClient interface {
	GetMysqlServices(cliConnection plugin.CliConnection) ([]MysqlService, error)
	GetStartedApps(cliConnection plugin.CliConnection) ([]GetAppsModel, error)
	OpenSshTunnel(cliConnection plugin.CliConnection, toService MysqlService, throughApp string, localPort int)
}

type MysqlService struct {
	Name     string
	AppName  string
	Hostname string
	Port     string
	DbName   string
	Username string
	Password string
}

type SdkApiClient struct {
	SshRunner  SshRunner
	PortWaiter PortWaiter
	HttpClient Http
}

func NewSdkApiClient() *SdkApiClient {
	return &SdkApiClient{
		SshRunner: new(CfSshRunner),
		PortWaiter: NewPortWaiter(),
		HttpClient: new(HttpWrapper),
	}
}

func (self *SdkApiClient) GetMysqlServices(cliConnection plugin.CliConnection) ([]MysqlService, error) {
	bindings, err := self.GetServiceBindings(cliConnection)
	if err != nil {
		return nil, err
	}

	instances, err := self.GetServiceInstances(cliConnection)
	if err != nil {
		return nil, err
	}

	return getAvailableServices(bindings, instances), nil
}

func (self *SdkApiClient) GetServiceBindings(cliConnection plugin.CliConnection) (*pluginResources.PaginatedServiceBindingResources, error) {
	bindingsResp, err := self.getFromCfApi("/v2/service_bindings", cliConnection)
	if err != nil {
		return nil, fmt.Errorf("Unable to call service bindings endpoint: %s", err)
	}

	return deserializeBindings(bindingsResp)
}

func (self *SdkApiClient) getFromCfApi(path string, cliConnection plugin.CliConnection) ([]byte, error) {
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

func deserializeBindings(bindingResponse []byte) (*pluginResources.PaginatedServiceBindingResources, error) {
	paginatedResources := new(pluginResources.PaginatedServiceBindingResources)
	err := json.Unmarshal(bindingResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service bindings: %s", err)
	}

	return paginatedResources, nil
}

func (self *SdkApiClient) GetServiceInstances(cliConnection plugin.CliConnection) (*resources.PaginatedServiceInstanceResources, error) {
	instanceResponse, err := self.getFromCfApi("/v2/service_instances", cliConnection)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve service instances: %s", err)
	}

	return deserializeInstances(instanceResponse)
}

func deserializeInstances(jsonResponse []byte) (*resources.PaginatedServiceInstanceResources, error) {
	paginatedResources := new(resources.PaginatedServiceInstanceResources)
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service instances: %s", err)
	}

	return paginatedResources, nil
}

func (self *SdkApiClient) GetStartedApps(cliConnection plugin.CliConnection) ([]GetAppsModel, error) {
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

func (self *SdkApiClient) OpenSshTunnel(cliConnection plugin.CliConnection, toService MysqlService, throughApp string, localPort int) {
	go self.SshRunner.OpenSshTunnel(cliConnection, toService, throughApp, localPort)

	self.PortWaiter.WaitUntilOpen(localPort)
}

func getAvailableServices(bindings *pluginResources.PaginatedServiceBindingResources, instances *resources.PaginatedServiceInstanceResources) []MysqlService {
	boundServiceCredentials := make(map[string]pluginResources.MysqlCredentials)

	for _, bindingResource := range (bindings.Resources) {
		guid := bindingResource.Entity.ServiceInstanceGUID
		credentials := bindingResource.Entity.Credentials

		boundServiceCredentials[guid] = credentials
	}

	services := make([]MysqlService, 0, len(bindings.Resources))

	for _, instanceResource := range (instances.Resources) {
		guid := instanceResource.Metadata.GUID
		name := instanceResource.Entity.Name

		credentials, serviceBound := boundServiceCredentials[guid]
		if serviceBound && strings.HasPrefix(credentials.Uri, "mysql://") {
			services = append(services, makeServiceModel(name, credentials))
		}
	}

	return services
}

func makeServiceModel(name string, credentials pluginResources.MysqlCredentials) MysqlService {
	return MysqlService{
		Name: name,
		Hostname: credentials.Hostname,
		Port: credentials.Port,
		DbName: credentials.DbName,
		Username: credentials.Username,
		Password: credentials.Password,
	}
}
