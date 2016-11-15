package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/andreasf/cf-mysql-plugin/cfmysql"
	"os"
	"fmt"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "This executable is a cf plugin. " +
			"Run `cf install-plugin %s` to install it\nand `cf mysql service-name` " +
			"to use it.\n",
			os.Args[0])
		os.Exit(1)
	}
	plugin.Start(cfmysql.NewPlugin())
}
