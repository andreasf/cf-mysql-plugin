package cfmysql

import (
	"net/http"
	"io/ioutil"
	"fmt"
)

//go:generate counterfeiter . Http
type Http interface {
	Get(endpoint string, access_token string) ([]byte, error)
}

type HttpWrapper struct {}

func (self *HttpWrapper) Get(url string, accessToken string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err)
	}

	request.Header.Add("Authorization", accessToken)

	client := new(http.Client)
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