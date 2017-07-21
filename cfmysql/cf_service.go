package cfmysql

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
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

type cfService struct {
	apiClient  ApiClient
	httpClient Http
	portWaiter PortWaiter
	sshRunner  SshRunner
}

func NewCfService(apiClient ApiClient, runner SshRunner, waiter PortWaiter, httpClient Http) *cfService {
	return &cfService{
		apiClient:  apiClient,
		sshRunner:  runner,
		portWaiter: waiter,
		httpClient: httpClient,
	}
}

type BindingResult struct {
	Bindings []pluginModels.ServiceBinding
	Err      error
}

func (self *cfService) GetMysqlServices(cliConnection plugin.CliConnection) ([]MysqlService, error) {
	bindingChan := make(chan BindingResult, 0)
	go func() {
		bindings, err := self.apiClient.GetServiceBindings(cliConnection)
		bindingChan <- BindingResult{Bindings: bindings, Err: err}
	}()

	instances, err := self.apiClient.GetServiceInstances(cliConnection)
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

func (self *cfService) GetStartedApps(cliConnection plugin.CliConnection) ([]sdkModels.GetAppsModel, error) {
	return self.apiClient.GetStartedApps(cliConnection)
}

func (self *cfService) OpenSshTunnel(cliConnection plugin.CliConnection, toService MysqlService, throughApp string, localPort int) {
	go self.sshRunner.OpenSshTunnel(cliConnection, toService, throughApp, localPort)

	self.portWaiter.WaitUntilOpen(localPort)
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
		Name:     name,
		Hostname: binding.Hostname,
		Port:     binding.Port,
		DbName:   binding.DbName,
		Username: binding.Username,
		Password: binding.Password,
	}
}
