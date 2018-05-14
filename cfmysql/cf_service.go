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
	OpenSshTunnel(cliConnection plugin.CliConnection, toService MysqlService, apps []sdkModels.GetAppsModel, localPort int)
	GetService(connection plugin.CliConnection, name string) (MysqlService, error)
}

func NewCfService(apiClient ApiClient, runner SshRunner, waiter PortWaiter, httpClient HttpWrapper, randWrapper RandWrapper) *cfService {
	return &cfService{
		apiClient:   apiClient,
		sshRunner:   runner,
		portWaiter:  waiter,
		httpClient:  httpClient,
		randWrapper: randWrapper,
	}
}

const SERVICE_KEY_NAME = "cf-mysql"

type MysqlService struct {
	Name     string
	Hostname string
	Port     string
	DbName   string
	Username string
	Password string
}

type cfService struct {
	apiClient   ApiClient
	httpClient  HttpWrapper
	portWaiter  PortWaiter
	sshRunner   SshRunner
	randWrapper RandWrapper
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

func (self *cfService) OpenSshTunnel(cliConnection plugin.CliConnection, toService MysqlService, apps []sdkModels.GetAppsModel, localPort int) {
	throughAppIndex := self.randWrapper.Intn(len(apps))
	throughApp := apps[throughAppIndex].Name
	go self.sshRunner.OpenSshTunnel(cliConnection, toService, throughApp, localPort)

	self.portWaiter.WaitUntilOpen(localPort)
}

func (self *cfService) GetService(connection plugin.CliConnection, name string) (MysqlService, error) {
	space, err := connection.GetCurrentSpace()
	if err != nil {
		return MysqlService{}, fmt.Errorf("unable to retrieve current space: %s", err)
	}

	instance, err := self.apiClient.GetService(connection, space.Guid, name)
	if err != nil {
		return MysqlService{}, fmt.Errorf("unable to retrieve metadata for service %s: %s", name, err)
	}

	serviceKey, found, err := self.apiClient.GetServiceKey(connection, instance.Guid, SERVICE_KEY_NAME)
	if err != nil {
		return MysqlService{}, fmt.Errorf("unable to retrieve service key: %s", err)
	}

	if found {
		return toServiceModel(name, serviceKey), nil
	}

	serviceKey, err = self.apiClient.CreateServiceKey(connection, instance.Guid, SERVICE_KEY_NAME)
	if err != nil {
		return MysqlService{}, fmt.Errorf("unable to create service key: %s", err)
	}

	return toServiceModel(name, serviceKey), nil
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

func toServiceModel(name string, serviceKey pluginModels.ServiceKey) MysqlService {
	return MysqlService{
		Name:     name,
		Hostname: serviceKey.Hostname,
		Port:     serviceKey.Port,
		DbName:   serviceKey.DbName,
		Username: serviceKey.Username,
		Password: serviceKey.Password,
	}
}
