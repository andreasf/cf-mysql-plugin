package cfmysql_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"fmt"
	"net/http"

	"bytes"
	"code.cloudfoundry.org/cli/cf/net/netfakes"
	"github.com/andreasf/cf-mysql-plugin/cfmysql"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("HttpWrapper", func() {

	Describe("Get", func() {
		Context("When SSL is disabled", func() {
			It("Configures an HTTP client without cert validation", func() {
				mockFactory := new(cfmysqlfakes.FakeHttpClientFactory)
				httpWrapper := cfmysql.NewHttpWrapper(mockFactory, new(netfakes.FakeRequestDumperInterface))

				mockFactory.NewClientReturns(new(http.Client))
				httpWrapper.Get("http://0.0.0.0/foo", "the-authorization-value", true)

				sslDisabled := mockFactory.NewClientArgsForCall(0)

				Expect(sslDisabled).To(BeTrue())
			})
		})

		Context("When SSL is enabled", func() {
			It("Does not disable cert validation", func() {
				mockFactory := new(cfmysqlfakes.FakeHttpClientFactory)
				httpWrapper := cfmysql.NewHttpWrapper(mockFactory, new(netfakes.FakeRequestDumperInterface))

				mockFactory.NewClientReturns(new(http.Client))
				httpWrapper.Get("http://0.0.0.0/foo", "the-authorization-value", false)

				sslDisabled := mockFactory.NewClientArgsForCall(0)

				Expect(sslDisabled).To(BeFalse())
			})
		})

		Context("When the request is successful", func() {
			It("Returns a response", func() {
				mockServer := ghttp.NewServer()
				mockServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, "response body"),
					ghttp.VerifyRequest("GET", "/the/expected/path"),
					ghttp.VerifyHeaderKV("Authorization", "the-authorization-value"),
				))

				httpWrapper := MakeHttp()

				response, err := httpWrapper.Get(mockServer.URL()+"/the/expected/path", "the-authorization-value", true)

				Expect(response).To(Equal([]byte("response body")))
				Expect(err).To(BeNil())

				mockServer.Close()
			})
		})

		Context("When the request returns an error", func() {
			It("Returns an error and no response", func() {
				httpWrapper := MakeHttp()

				response, err := httpWrapper.Get("http://localhost:52719/no/server/here", "foo", true)

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

				httpWrapper := MakeHttp()
				response, err := httpWrapper.Get(mockServer.URL()+"/need/coffee", "foo", true)

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

				httpWrapper := MakeHttp()
				response, err := httpWrapper.Get(mockServer.URL()+"/this/shall/crash", "foo", true)

				Expect(err).To(Equal(fmt.Errorf("HTTP status 500 accessing %s/this/shall/crash", mockServer.URL())))
				Expect(response).To(BeNil())

				mockServer.Close()
			})
		})
	})

	Describe("Post", func() {
		It("Sends a body and sets the content type to application/json", func() {
			headers := make(http.Header)
			headers.Add("Content-Type", "application/json")
			body := []byte("hello there")

			mockServer := ghttp.NewServer()
			mockServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, "ok"),
				ghttp.VerifyRequest("POST", "/path"),
				ghttp.VerifyBody(body),
				ghttp.VerifyHeader(headers),
			))

			dumper := new(netfakes.FakeRequestDumperInterface)
			httpWrapper := cfmysql.NewHttpWrapper(cfmysql.NewHttpClientFactory(), dumper)

			response, err := httpWrapper.Post(mockServer.URL()+"/path", bytes.NewBuffer(body), "access-token", true)

			Expect(err).To(BeNil())
			Expect(response).To(Equal([]byte("ok")))

			Expect(dumper.DumpRequestCallCount()).To(Equal(1))
			Expect(dumper.DumpResponseCallCount()).To(Equal(1))

			mockServer.Close()
		})

		Context("When SSL is disabled", func() {
			It("Configures an HTTP client without cert validation", func() {
				mockFactory := new(cfmysqlfakes.FakeHttpClientFactory)
				httpWrapper := cfmysql.NewHttpWrapper(mockFactory, new(netfakes.FakeRequestDumperInterface))

				mockFactory.NewClientReturns(new(http.Client))
				httpWrapper.Post("http://0.0.0.0/foo", bytes.NewBufferString(""), "the-authorization-value", true)

				sslDisabled := mockFactory.NewClientArgsForCall(0)

				Expect(sslDisabled).To(BeTrue())
			})
		})

		Context("When SSL is enabled", func() {
			It("Does not disable cert validation", func() {
				mockFactory := new(cfmysqlfakes.FakeHttpClientFactory)
				httpWrapper := cfmysql.NewHttpWrapper(mockFactory, new(netfakes.FakeRequestDumperInterface))

				mockFactory.NewClientReturns(new(http.Client))
				httpWrapper.Post("http://0.0.0.0/foo", bytes.NewBufferString(""), "the-authorization-value", false)

				sslDisabled := mockFactory.NewClientArgsForCall(0)

				Expect(sslDisabled).To(BeFalse())
			})
		})
	})
})

func MakeHttp() cfmysql.HttpWrapper {
	return cfmysql.NewHttpWrapper(cfmysql.NewHttpClientFactory(), new(netfakes.FakeRequestDumperInterface))
}
