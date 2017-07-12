package cfmysql_test

import (
	. "github.com/andreasf/cf-mysql-plugin/cfmysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("HttpClientFactory", func() {
	Context("When SSL is disabled", func() {
		It("Enables skipping certificates", func() {
			factory := NewHttpClientFactory()

			client := factory.NewClient(true)
			transport := client.Transport.(*http.Transport)

			Expect(transport.TLSClientConfig.InsecureSkipVerify).To(BeTrue())
		})
	})

	Context("When SSL is enabled", func() {
		It("Checks certificates", func() {
			factory := NewHttpClientFactory()

			client := factory.NewClient(false)
			transport := client.Transport.(*http.Transport)

			Expect(transport.TLSClientConfig.InsecureSkipVerify).To(BeFalse())
		})
	})
})
