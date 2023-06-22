package cfmysql_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfmysql(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfmysql Suite")
}
