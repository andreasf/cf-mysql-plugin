package resources

import (
	"code.cloudfoundry.org/cli/cf/api/resources"
	"encoding/json"
	"fmt"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/models"
	"strconv"
	"strings"
)

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
	Tls      TlsResource     `json:"tls"`
}

type TlsResource struct {
	Cert TlsCertResource `json:"cert"`
}

type TlsCertResource struct {
	Ca string `json:"ca"`
}

type PaginatedServiceInstanceResources struct {
	TotalResults int    `json:"total_results"`
	NextUrl      string `json:"next_url"`
	Resources    []ServiceInstanceResource
}

type ServiceInstanceResource struct {
	resources.Resource
	Entity ServiceInstanceEntity
}

type PaginatedServiceKeyResources struct {
	TotalResults int    `json:"total_results"`
	NextUrl      string `json:"next_url"`
	Resources    []ServiceKeyResource
}

type ServiceKeyResource struct {
	resources.Resource
	Entity ServiceKeyEntity
}

type ServiceKeyEntity struct {
	ServiceInstanceName string `json:"name"`
	ServiceInstanceGuid string `json:"service_instance_guid"`
	Credentials         MysqlCredentials
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
	var convertedModels []models.ServiceInstance

	for _, resource := range self.Resources {
		model := models.ServiceInstance{}
		model.Guid = resource.Metadata.GUID
		model.Name = resource.Entity.Name

		pathParts := strings.Split(resource.Entity.SpaceUrl, "/")
		model.SpaceGuid = pathParts[len(pathParts)-1]

		convertedModels = append(convertedModels, model)
	}

	return convertedModels
}

func (self *PaginatedServiceKeyResources) ToModel() ([]models.ServiceKey, error) {
	var convertedModels []models.ServiceKey

	for _, resource := range self.Resources {
		model, err := resource.ToModel()
		if err != nil {
			return nil, err
		}

		convertedModels = append(convertedModels, model)
	}

	return convertedModels, nil
}

func (self *ServiceKeyResource) ToModel() (models.ServiceKey, error) {
	var port string

	if len(self.Entity.Credentials.RawPort) > 0 {
		var portInt int
		var portString string

		err := json.Unmarshal(self.Entity.Credentials.RawPort, &portString)
		if err != nil {
			err = json.Unmarshal(self.Entity.Credentials.RawPort, &portInt)
			if err != nil {
				return models.ServiceKey{}, fmt.Errorf("unable to deserialize port in service key: '%s'", string(self.Entity.Credentials.RawPort))
			}
			portString = strconv.Itoa(portInt)
		}
		port = portString
	}

	return models.ServiceKey{
		ServiceInstanceGuid: self.Entity.ServiceInstanceGuid,
		Uri:                 self.Entity.Credentials.Uri,
		DbName:              self.Entity.Credentials.DbName,
		Hostname:            self.Entity.Credentials.Hostname,
		Port:                port,
		Username:            self.Entity.Credentials.Username,
		Password:            self.Entity.Credentials.Password,
		CaCert:              self.Entity.Credentials.Tls.Cert.Ca,
	}, nil
}
