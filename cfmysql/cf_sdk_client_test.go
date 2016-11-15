package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/test_resources"
	"fmt"
)

var _ = Describe("CfSdkClient", func() {
	Context("Retrieving available MySQL services", func() {
		var sdkClient *SdkCfClient
		var cliConnection *pluginfakes.FakeCliConnection

		BeforeEach(func() {
			cliConnection = new(pluginfakes.FakeCliConnection)
			cliConnection.CliCommandWithoutTerminalOutputStub = mockCfCurl
			sdkClient = new(SdkCfClient)
		})

		It("Gets a list of bindings", func() {
			paginatedResources, err := sdkClient.GetServiceBindings(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(3))
		})

		It("Gets a list of instances", func() {
			paginatedResources, err := sdkClient.GetServiceInstances(cliConnection)

			Expect(err).To(BeNil())
			Expect(paginatedResources.Resources).To(HaveLen(4))
		})

		It("Gets a list of service instances and bindings", func() {
			services, err := sdkClient.GetMysqlServices(cliConnection)

			Expect(err).To(BeNil())
			Expect(services).To(HaveLen(2))

			Expect(services[0]).To(Equal(MysqlService{
				Name: "database-a",
				AppName: "",
				Hostname: "database-a.host",
				Port: "3306",
				DbName: "dbname-a",
				Username: "username-a",
				Password: "password-a",
			}))

			Expect(services[1]).To(Equal(MysqlService{
				Name: "database-b",
				AppName: "",
				Hostname: "database-b.host",
				Port: "3307",
				DbName: "dbname-b",
				Username: "username-b",
				Password: "password-b",
			}))

			Expect(cliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(2))
			Expect(cliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)).To(Equal([]string{"curl", "/v2/service_bindings"}))
			Expect(cliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)).To(Equal([]string{"curl", "/v2/service_instances"}))
		})
	})
})

func mockCfCurl(args... string) ([]string, error) {
	if SliceEquals(args, []string{"curl", "/v2/service_bindings"}) {
		return test_resources.LoadResourceLines("test_resources/service_bindings.json"), nil
	}
	if SliceEquals(args, []string{"curl", "/v2/service_instances"}) {
		return test_resources.LoadResourceLines("test_resources/service_instances.json"), nil
	}

	return nil, fmt.Errorf("No mock defined for %s", args)
}

func SliceEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
