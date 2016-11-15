package resources

import "code.cloudfoundry.org/cli/cf/api/resources"

type PaginatedServiceBindingResources struct {
	TotalResults int `json:"total_results"`
	Resources    []ServiceBindingResource
}

type ServiceBindingResource struct {
	resources.Resource
	Entity ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGUID string `json:"app_guid"`
	ServiceInstanceGUID string `json:"service_instance_guid"`
	Credentials MysqlCredentials `json:"credentials"`
}

type MysqlCredentials struct {
	Uri string `json:"uri"`
	DbName string `json:"name"`
	Hostname string `json:"hostname"`
	Port string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}
