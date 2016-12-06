package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
	"strings"
	sdkModels "code.cloudfoundry.org/cli/plugin/models"
	pluginModels "github.com/andreasf/cf-mysql-plugin/cfmysql/models"
)

//go:generate counterfeiter . CfService
type CfService interface {
	GetMysqlServices(cliConnection plugin.CliConnection) ([]MysqlService, error)
	GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error)
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

type CfServiceImpl struct {
	ApiClient  ApiClient
	HttpClient Http
	PortWaiter PortWaiter
	SshRunner  SshRunner
}

func NewCfService() *CfServiceImpl {
	return &CfServiceImpl{
		ApiClient: NewApiClient(),
		SshRunner: new(CfSshRunner),
		PortWaiter: NewPortWaiter(),
		HttpClient: new(HttpWrapper),
	}
}

type BindingResult struct {
	Bindings []pluginModels.ServiceBinding
	Err      error
}

func (self *CfServiceImpl) GetMysqlServices(cliConnection plugin.CliConnection) ([]MysqlService, error) {
	bindingChan := make(chan BindingResult, 0)
	go func() {
		bindings, err := self.ApiClient.GetServiceBindings(cliConnection)
		bindingChan <- BindingResult{Bindings: bindings, Err: err}
	}()

	instances, err := self.ApiClient.GetServiceInstances(cliConnection)
	if err != nil {
		return nil, err
	}

	bindingResult := <-bindingChan
	if bindingResult.Err != nil {
		return nil, bindingResult.Err
	}

	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving current space: %s", err)
	}

	return getAvailableServices(bindingResult.Bindings, instances, space.Guid), nil
}

func (self *CfServiceImpl) getFromCfApi(path string, cliConnection plugin.CliConnection) ([]byte, error) {
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

func (self *CfServiceImpl) GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error) {
	return self.ApiClient.GetStartedApps(cliConnection)
}

func (self *CfServiceImpl) OpenSshTunnel(cliConnection plugin.CliConnection, toService MysqlService, throughApp string, localPort int) {
	go self.SshRunner.OpenSshTunnel(cliConnection, toService, throughApp, localPort)

	self.PortWaiter.WaitUntilOpen(localPort)
}

func getAvailableServices(bindings []pluginModels.ServiceBinding, instances []pluginModels.ServiceInstance, spaceGuid string) []MysqlService {
	boundServiceCredentials := make(map[string]pluginModels.ServiceBinding)

	for _, binding := range bindings {
		boundServiceCredentials[binding.ServiceInstanceGuid] = binding
	}

	services := make([]MysqlService, 0, len(bindings))

	for _, instance := range instances {
		guid := instance.Guid
		name := instance.Name

		binding, serviceBound := boundServiceCredentials[guid]
		if instance.SpaceGuid == spaceGuid && serviceBound && strings.HasPrefix(binding.Uri, "mysql://") {
			services = append(services, makeServiceModel(name, binding))
		}
	}

	return services
}

func makeServiceModel(name string, binding pluginModels.ServiceBinding) MysqlService {
	return MysqlService{
		Name: name,
		Hostname: binding.Hostname,
		Port: binding.Port,
		DbName: binding.DbName,
		Username: binding.Username,
		Password: binding.Password,
	}
}
