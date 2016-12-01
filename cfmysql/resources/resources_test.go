package resources_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/test_resources"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"encoding/json"
)

var _ = Describe("Resources", func() {
	Context("Deserializing JSON", func() {
		Context("Service instances", func() {
			It("can get service instance names and guids", func() {
				paginatedResources := new(resources.PaginatedServiceInstanceResources)
				err := json.Unmarshal(test_resources.LoadResource("../test_resources/service_instances.json"), paginatedResources)

				Expect(err).To(BeNil())
				Expect(paginatedResources.Resources).To(HaveLen(4))

				Expect(paginatedResources.Resources[0].Entity.Name).To(Equal("database-a"))
				Expect(paginatedResources.Resources[0].Metadata.GUID).To(Equal("service-instance-guid-a"))

				Expect(paginatedResources.Resources[1].Entity.Name).To(Equal("database-b"))
				Expect(paginatedResources.Resources[1].Metadata.GUID).To(Equal("service-instance-guid-b"))
			})
		})


		Context("Service bindings", func() {
			It("can get credentials", func() {
				paginatedResources := new(PaginatedServiceBindingResources)
				err := json.Unmarshal(test_resources.LoadResource("../test_resources/service_bindings.json"), paginatedResources)

				Expect(err).To(BeNil())
				Expect(paginatedResources.Resources).To(HaveLen(4))

				Expect(paginatedResources.Resources[0].Entity.ServiceInstanceGUID).To(Equal("service-instance-guid-a"))
				Expect(paginatedResources.Resources[0].Entity.Credentials.Uri).To(Equal("mysql://username-a:password-a@database-a.host:3306/dbname-a?reconnect=true"))
				Expect(paginatedResources.Resources[0].Entity.Credentials.DbName).To(Equal("dbname-a"))
				Expect(paginatedResources.Resources[0].Entity.Credentials.Hostname).To(Equal("database-a.host"))
				Expect(paginatedResources.Resources[0].Entity.Credentials.Username).To(Equal("username-a"))
				Expect(paginatedResources.Resources[0].Entity.Credentials.Password).To(Equal("password-a"))

				Expect(paginatedResources.Resources[1].Entity.ServiceInstanceGUID).To(Equal("service-instance-guid-b"))

				var portString string
				err = json.Unmarshal(paginatedResources.Resources[0].Entity.Credentials.RawPort, &portString)
				Expect(err).To(BeNil())
				Expect(portString).To(Equal("3306"))

				var portInt int
				err = json.Unmarshal(paginatedResources.Resources[3].Entity.Credentials.RawPort, &portInt)
				Expect(err).To(BeNil())
				Expect(portInt).To(Equal(54321))
			})
		})
	})
})
