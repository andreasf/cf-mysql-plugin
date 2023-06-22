package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os/exec"
)

var _ = Describe("Main", func() {
	It("Signals failure and prints an error message if run as a standalone app", func() {
		buildCmd := exec.Command("go", "build")
		err := buildCmd.Run()
		Expect(err).To(BeNil())

		runCmd := exec.Command("./cf-mysql-plugin")
		output, err := runCmd.CombinedOutput()

		Expect(err).ToNot(BeNil())
		Expect(string(output)).To(HavePrefix("This executable is a cf plugin."))
	})
})
