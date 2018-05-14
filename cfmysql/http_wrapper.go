package cfmysql

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"io"
)

//go:generate counterfeiter . HttpWrapper
type HttpWrapper interface {
	Get(endpoint string, accessToken string, skipSsl bool) ([]byte, error)
	Post(url string, body io.Reader, accessToken string, sslDisabled bool) ([]byte, error)
}

func NewHttpWrapper(factory HttpClientFactory) HttpWrapper {
	return &httpWrapper{
		httpClientFactory: factory,
	}
}

type httpWrapper struct {
	httpClientFactory HttpClientFactory
}

func (self *httpWrapper) Get(url string, accessToken string, sslDisabled bool) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	return self.do(request, accessToken, sslDisabled)
}

func (self *httpWrapper) Post(url string, body io.Reader, accessToken string, sslDisabled bool) ([]byte, error) {
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	request.Header.Add("Content-Type", "application/json")

	return self.do(request, accessToken, sslDisabled)
}

func (self *httpWrapper) do(request *http.Request, accessToken string, sslDisabled bool) ([]byte, error) {
	request.Header.Add("Authorization", accessToken)

	client := self.httpClientFactory.NewClient(sslDisabled)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if !isSuccessCode(response.StatusCode) {
		return nil, fmt.Errorf("HTTP status %d accessing %s", response.StatusCode, request.URL.String())
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func isSuccessCode(statusCode int) bool {
	switch statusCode / 100 {
	case 4:
		fallthrough
	case 5:
		return false
	}

	return true
}
