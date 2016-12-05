package resources_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/test_resources"
	"encoding/json"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/models"
)

var _ = Describe("Resources", func() {
	Describe("Service instances", func() {
		Context("Deserializing JSON", func() {
			It("can get service instance names and guids", func() {
				paginatedResources := new(PaginatedServiceInstanceResources)
				err := json.Unmarshal(test_resources.LoadResource("../test_resources/service_instances.json"), paginatedResources)

				Expect(err).To(BeNil())
				Expect(paginatedResources.Resources).To(HaveLen(4))

				Expect(paginatedResources.Resources[0].Entity.Name).To(Equal("database-a"))
				Expect(paginatedResources.Resources[0].Metadata.GUID).To(Equal("service-instance-guid-a"))
				Expect(paginatedResources.Resources[0].Entity.SpaceUrl).To(Equal("/v2/spaces/space-guid"))

				Expect(paginatedResources.Resources[1].Entity.Name).To(Equal("database-b"))
				Expect(paginatedResources.Resources[1].Metadata.GUID).To(Equal("service-instance-guid-b"))
				Expect(paginatedResources.Resources[1].Entity.SpaceUrl).To(Equal("/v2/spaces/space-guid"))
			})
		})

		Context("Converting to models", func() {
			It("Converts very nicely", func() {
				resourceInstances := &PaginatedServiceInstanceResources{
					Resources: []ServiceInstanceResource{
						{
							Resource: resources.Resource{
								Metadata: resources.Metadata{
									GUID: "fine-guid",
								},
							},
							Entity: ServiceInstanceEntity{
								Name: "outstanding-service-name",
								SpaceUrl: "/v2/spaces/outer-space-guid",
							},
						},
						{
							Resource: resources.Resource{
								Metadata: resources.Metadata{
									GUID: "better-guid",
								},
							},
							Entity: ServiceInstanceEntity{
								Name: "best-service-name",
								SpaceUrl: "/v2/spaces/inner-space-guid",
							},
						},
					},
				}

				instanceModels := resourceInstances.ToModel()

				Expect(instanceModels).To(HaveLen(2))
				Expect(instanceModels[0]).To(Equal(models.ServiceInstance{
					Name: "outstanding-service-name",
					Guid: "fine-guid",
					SpaceGuid: "outer-space-guid",
				}))
				Expect(instanceModels[1]).To(Equal(models.ServiceInstance{
					Name: "best-service-name",
					Guid: "better-guid",
					SpaceGuid: "inner-space-guid",
				}))
			})
		})
	})

	Describe("Service bindings", func() {
		Context("Deserializing JSON", func() {
			It("can get credentials", func() {
				paginatedResources := new(PaginatedServiceBindingResources)
				err := json.Unmarshal(test_resources.LoadResource("../test_resources/service_bindings.json"), paginatedResources)

				Expect(err).To(BeNil())
				Expect(paginatedResources.Resources).To(HaveLen(5))

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