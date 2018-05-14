package test_resources

import (
	"io/ioutil"
)

func LoadResource(filename string) []byte {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return contents
}
