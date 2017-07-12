package cfmysql

import (
	"net/http"
	"crypto/tls"
)

//go:generate counterfeiter . HttpClientFactory
type HttpClientFactory interface {
	NewClient(sslDisabled bool) *http.Client
}

type httpClientFactory struct {}

func NewHttpClientFactory() HttpClientFactory {
	return new(httpClientFactory)
}

func (self *httpClientFactory) NewClient(sslDisabled bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: sslDisabled},
	}
	return &http.Client{Transport: transport}
}