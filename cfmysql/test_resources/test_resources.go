package test_resources

import (
	"io/ioutil"
	"strings"
)

func LoadResource(filename string) []byte {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return contents
}

func LoadResourceLines(filename string) []string {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return strings.Split(string(contents), "\n")
}
