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
}

func NewSdkApiClient() *SdkApiClient {
	return &SdkApiClient{
		SshRunner: new(CfSshRunner),
		PortWaiter: NewPortWaiter(),
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
	bindingLines, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v2/service_bindings")
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve service bindings: %s", err)
	}
	return deserializeBindings(bindingLines)
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

func deserializeBindings(bindingLines []string) (*pluginResources.PaginatedServiceBindingResources, error) {
	paginatedResources := new(pluginResources.PaginatedServiceBindingResources)
	jsonResponse := []byte(strings.Join(bindingLines, "\n"))
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service bindings: %s", err)
	}

	return paginatedResources, nil
}

func (self *SdkApiClient) GetServiceInstances(cliConnection plugin.CliConnection) (*resources.PaginatedServiceInstanceResources, error) {
	instanceLines, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v2/service_instances")
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve service instances: %s", err)
	}

	return deserializeInstances(instanceLines)
}

func deserializeInstances(instanceLines []string) (*resources.PaginatedServiceInstanceResources, error) {
	paginatedResources := new(resources.PaginatedServiceInstanceResources)
	jsonResponse := []byte(strings.Join(instanceLines, "\n"))
	err := json.Unmarshal(jsonResponse, paginatedResources)

	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize service instances: %s", err)
	}

	return paginatedResources, nil
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
