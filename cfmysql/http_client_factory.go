package cfmysql

import (
	"crypto/tls"
	"net/http"
)

//go:generate counterfeiter . HttpClientFactory
type HttpClientFactory interface {
	NewClient(sslDisabled bool) *http.Client
}

func NewHttpClientFactory() HttpClientFactory {
	return new(httpClientFactory)
}

type httpClientFactory struct{}

func (self *httpClientFactory) NewClient(sslDisabled bool) *http.Client {
	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: sslDisabled},
	}
	return &http.Client{Transport: transport}
}
