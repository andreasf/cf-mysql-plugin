package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
)

type MysqlPlugin struct{}

func (c *MysqlPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "mysql",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "mysql",
				HelpText: "Connect to a MySQL database service",
				UsageDetails: plugin.Usage{
					Usage: "mysql\n   cf mysql <service-name>",
				},
			},
		},
	}
}

func (c *MysqlPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "mysql" {
		fmt.Println("Running mysql")
	}
}
