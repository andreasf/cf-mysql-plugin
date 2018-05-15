package cfmysql_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/andreasf/cf-mysql-plugin/cfmysql"
	"github.com/andreasf/cf-mysql-plugin/cfmysql/cfmysqlfakes"
	"net/http"
	"regexp"
	"github.com/onsi/gomega/gbytes"
	"code.cloudfoundry.org/cli/cf/terminal"
	"net/url"
)

var _ = Describe("RequestDumper", func() {
	expectedRequest := "\nREQUEST: [timestamp]\nBARK /such/test HTTP/1.1\r\nHost: wow\r\n\r\n\n"
	expectedResponse := "\nRESPONSE: [timestamp]\nHTTP/1.1 200 BORK\r\nContent-Length: 0\r\n\r\n\n"

	var response http.Response
	var request http.Request

	BeforeEach(func() {
		url, err := url.Parse("bark://wow/such/test")
		Expect(err).To(BeNil())

		request = http.Request{
			Method: "BARK",
			URL: url,
			Proto: "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		}

		response = http.Response{
			Status: "BORK",
			StatusCode: 200,
			Proto: "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		}
	})

	Describe("DumpRequest", func() {
		Context("When CF_TRACE is set", func() {
			It("Dumps the request to the writer", func() {
				buffer := gbytes.NewBuffer()
				osWrapper := new(cfmysqlfakes.FakeOsWrapper)
				osWrapper.LookupEnvReturns("1", true)
				dumper := cfmysql.NewRequestDumper(osWrapper, buffer)

				dumper.DumpRequest(&request)

				requestString := removeTimestampAndColors(buffer.Contents())
				Expect(requestString).To(Equal(expectedRequest))
			})
		})

		Context("When CF_TRACE is not set", func() {
			It("Does not dump requests", func() {
				buffer := gbytes.NewBuffer()
				osWrapper := new(cfmysqlfakes.FakeOsWrapper)
				osWrapper.LookupEnvReturns("", false)
				dumper := cfmysql.NewRequestDumper(osWrapper, buffer)

				dumper.DumpRequest(&request)

				Expect(buffer.Contents()).To(BeEmpty())
			})
		})
	})

	Describe("DumpResponse", func() {
		Context("When CF_TRACE is set", func() {
			It("Dumps the response to the writer", func() {
				buffer := gbytes.NewBuffer()
				osWrapper := new(cfmysqlfakes.FakeOsWrapper)
				osWrapper.LookupEnvReturns("1", true)
				dumper := cfmysql.NewRequestDumper(osWrapper, buffer)

				dumper.DumpResponse(&response)

				responseString := removeTimestampAndColors(buffer.Contents())
				Expect(responseString).To(Equal(expectedResponse))
			})
		})

		Context("When CF_TRACE is not set", func() {
			It("Does not dump responses", func() {
				buffer := gbytes.NewBuffer()
				osWrapper := new(cfmysqlfakes.FakeOsWrapper)
				osWrapper.LookupEnvReturns("", false)
				dumper := cfmysql.NewRequestDumper(osWrapper, buffer)

				dumper.DumpResponse(&response)

				Expect(buffer.Contents()).To(BeEmpty())
			})
		})

	})
})

func removeTimestampAndColors(requestOrResponse []byte) string {
	re, err := regexp.Compile("\\[[^\\]]*\\]")
	Expect(err).To(BeNil())

	withoutColors := terminal.Decolorize(string(requestOrResponse))

	withoutTimestamp := string(re.ReplaceAll([]byte(withoutColors), []byte("[timestamp]")))

	return withoutTimestamp
}