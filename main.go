package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/andreasf/cf-mysql-plugin/cfmysql"
	"os"
	"fmt"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "This executable is a cf plugin. "+
			"Run `cf install-plugin %s` to install it\nand `cf mysql service-name` "+
			"to use it.\n",
			os.Args[0])
		os.Exit(1)
	}

	mysqlPlugin := newPlugin()
	plugin.Start(mysqlPlugin)

	os.Exit(mysqlPlugin.GetExitCode())
}

func newPlugin() *cfmysql.MysqlPlugin {
	httpClientFactory := cfmysql.NewHttpClientFactory()
	http := cfmysql.NewHttp(httpClientFactory)
	apiClient := cfmysql.NewApiClient(http)

	sshRunner := new(cfmysql.CfSshRunner)
	netWrapper := new(cfmysql.NetWrapper)
	waiter := cfmysql.NewPortWaiter(netWrapper)
	cfService := cfmysql.NewCfService(apiClient, sshRunner, waiter, http)

	execWrapper := new(cfmysql.ExecWrapper)
	runner := cfmysql.NewMysqlRunner(execWrapper)

	portFinder := new(cfmysql.FreePortFinder)

	return cfmysql.NewPlugin(cfmysql.PluginConf{
		In:          os.Stdin,
		Out:         os.Stdout,
		Err:         os.Stderr,
		CfService:   cfService,
		PortFinder:  portFinder,
		MysqlRunner: runner,
	})
}
