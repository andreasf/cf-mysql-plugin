package cfmysql_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
	"github.com/andreasf/cf-mysql-plugin/cfmysql"
	"net/http"
	"fmt"
)

var _ = Describe("HttpWrapper", func() {
	Context("When the request is successful", func() {
		It("Returns a response", func() {
			mockServer := ghttp.NewServer()
			mockServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, "response body"),
				ghttp.VerifyRequest("GET", "/the/expected/path"),
				ghttp.VerifyHeaderKV("Authorization", "the-authorization-value"),
			))

			httpWrapper := new(cfmysql.HttpWrapper)

			response, err := httpWrapper.Get(mockServer.URL() + "/the/expected/path", "the-authorization-value")

			Expect(response).To(Equal([]byte("response body")))
			Expect(err).To(BeNil())

			mockServer.Close()
		})
	})

	Context("When the request returns an error", func() {
		It("Returns an error and no response", func() {
			httpWrapper := new(cfmysql.HttpWrapper)

			response, err := httpWrapper.Get("http://localhost:52719/no/server/here", "foo")

			Expect(err).To(Not(BeNil()))
			Expect(response).To(BeNil())
		})
	})

	Context("When the response returns a 4xx response", func() {
		It("Returns an error and no response", func() {
			mockServer := ghttp.NewServer()
			mockServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusTeapot, "your mistake"),
				ghttp.VerifyRequest("GET", "/need/coffee")),
			)

			httpWrapper := new(cfmysql.HttpWrapper)
			response, err := httpWrapper.Get(mockServer.URL() + "/need/coffee", "foo")

			Expect(err).To(Equal(fmt.Errorf("HTTP status 418 accessing %s/need/coffee", mockServer.URL())))
			Expect(response).To(BeNil())

			mockServer.Close()
		})
	})

	Context("When the response returns a 5xx response", func() {
		It("Returns an error and no response", func() {
			mockServer := ghttp.NewServer()
			mockServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusInternalServerError, "my mistake"),
				ghttp.VerifyRequest("GET", "/this/shall/crash")),
			)

			httpWrapper := new(cfmysql.HttpWrapper)
			response, err := httpWrapper.Get(mockServer.URL() + "/this/shall/crash", "foo")

			Expect(err).To(Equal(fmt.Errorf("HTTP status 500 accessing %s/this/shall/crash", mockServer.URL())))
			Expect(response).To(BeNil())

			mockServer.Close()
		})
	})
})
