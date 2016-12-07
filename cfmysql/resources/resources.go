package resources

import (
	"code.cloudfoundry.org/cli/cf/api/resources"
	"encoding/json"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/models"
	"strings"
	"strconv"
	"fmt"
)

type PaginatedServiceBindingResources struct {
	TotalResults int `json:"total_results"`
	NextUrl      string `json:"next_url"`
	Resources    []ServiceBindingResource
}

type ServiceBindingResource struct {
	resources.Resource
	Entity ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGUID             string           `json:"app_guid"`
	ServiceInstanceGUID string           `json:"service_instance_guid"`
	Credentials         MysqlCredentials `json:"credentials"`
}

type MysqlCredentials struct {
	Uri      string `json:"uri"`
	DbName   string `json:"name"`
	Hostname string `json:"hostname"`
	Port     string
	RawPort  json.RawMessage `json:"port"`
	Username string          `json:"username"`
	Password string          `json:"password"`
}

type PaginatedServiceInstanceResources struct {
	TotalResults int `json:"total_results"`
	NextUrl      string `json:"next_url"`
	Resources    []ServiceInstanceResource
}

type ServiceInstanceResource struct {
	resources.Resource
	Entity ServiceInstanceEntity
}

type ServiceInstanceEntity struct {
	Name            string                         `json:"name"`
	DashboardURL    string                         `json:"dashboard_url"`
	Tags            []string                       `json:"tags"`
	ServiceBindings []ServiceBindingResource       `json:"service_bindings"`
	ServiceKeys     []resources.ServiceKeyResource `json:"service_keys"`
	ServicePlan     resources.ServicePlanResource  `json:"service_plan"`
	LastOperation   resources.LastOperation        `json:"last_operation"`
	SpaceUrl        string                         `json:"space_url"`
}

func (self *PaginatedServiceInstanceResources) ToModel() []models.ServiceInstance {
	convertedModels := []models.ServiceInstance{}

	for _, resource := range self.Resources {
		model := models.ServiceInstance{}
		model.Guid = resource.Metadata.GUID
		model.Name = resource.Entity.Name

		pathParts := strings.Split(resource.Entity.SpaceUrl, "/")
		model.SpaceGuid = pathParts[len(pathParts) - 1]

		convertedModels = append(convertedModels, model)
	}

	return convertedModels
}

func (self *PaginatedServiceBindingResources) ToModel() ([]models.ServiceBinding, error) {
	convertedModels := []models.ServiceBinding{}

	for _, resource := range self.Resources {
		var port string

		if len(resource.Entity.Credentials.RawPort) > 0 {
			var portInt int
			var portString string

			err := json.Unmarshal(resource.Entity.Credentials.RawPort, &portString)
			if err != nil {
				err = json.Unmarshal(resource.Entity.Credentials.RawPort, &portInt)
				if err != nil {
					return nil, fmt.Errorf("Unable to deserialize port in service binding: '%s'", string(resource.Entity.Credentials.RawPort))
				}
				portString = strconv.Itoa(portInt)
			}
			port = portString
		}

		binding := models.ServiceBinding{
			ServiceInstanceGuid: resource.Entity.ServiceInstanceGUID,
			Uri: resource.Entity.Credentials.Uri,
			DbName: resource.Entity.Credentials.DbName,
			Hostname: resource.Entity.Credentials.Hostname,
			Port: port,
			Username: resource.Entity.Credentials.Username,
			Password: resource.Entity.Credentials.Password,
		}

		convertedModels = append(convertedModels, binding)
	}

	return convertedModels, nil
}