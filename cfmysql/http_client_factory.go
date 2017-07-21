package cfmysql

import (
	"net/http"
	"crypto/tls"
)

//go:generate counterfeiter . HttpClientFactory
type HttpClientFactory interface {
	NewClient(sslDisabled bool) *http.Client
}

func NewHttpClientFactory() HttpClientFactory {
	return new(httpClientFactory)
}

type httpClientFactory struct {}

func (self *httpClientFactory) NewClient(sslDisabled bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: sslDisabled},
	}
	return &http.Client{Transport: transport}
}