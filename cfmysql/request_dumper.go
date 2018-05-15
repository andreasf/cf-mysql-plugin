package cfmysql

import (
	"code.cloudfoundry.org/cli/cf/net"
	"io"
	"net/http"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/trace"
)

type conditionalRequestDumper struct {
	requestDumper net.RequestDumperInterface
	traceRequests bool
}

func NewRequestDumper(osWrapper OsWrapper, writer io.Writer) *conditionalRequestDumper {
	i18n.T = i18n.Init(&DefaultLocaleReader{})

	_, traceRequests := osWrapper.LookupEnv("CF_TRACE")

	printer := trace.NewWriterPrinter(writer, true)
	dumper := net.NewRequestDumper(printer)

	return &conditionalRequestDumper{
		requestDumper: dumper,
		traceRequests: traceRequests,
	}
}

func (self *conditionalRequestDumper) DumpRequest(request *http.Request) {
	if self.traceRequests {
		self.requestDumper.DumpRequest(request)
	}
}

func (self *conditionalRequestDumper) DumpResponse(response *http.Response) {
	if self.traceRequests {
		self.requestDumper.DumpResponse(response)
	}
}

type DefaultLocaleReader struct{}

func (self *DefaultLocaleReader) Locale() string {
	return ""
}
