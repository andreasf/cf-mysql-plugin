package cfmysql

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

//go:generate counterfeiter . Http
type Http interface {
	Get(endpoint string, access_token string, skipSsl bool) ([]byte, error)
}

type HttpWrapper struct {
	HttpClientFactory HttpClientFactory
}

func NewHttp(factory HttpClientFactory) Http {
	return &HttpWrapper{
		HttpClientFactory: factory,
	}
}

func (self *HttpWrapper) Get(url string, accessToken string, sslDisabled bool) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err)
	}

	request.Header.Add("Authorization", accessToken)
	client := self.HttpClientFactory.NewClient(sslDisabled)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if !isSuccessCode(response.StatusCode) {
		return nil, fmt.Errorf("HTTP status %d accessing %s", response.StatusCode, url)
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
