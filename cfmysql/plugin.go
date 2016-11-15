package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"io"
	"fmt"
	"os"
)

type MysqlPlugin struct {
	In        io.Reader
	Out       io.Writer
	Err       io.Writer
	SdkClient CfClient
	ExitCode  int
}

func (self *MysqlPlugin) GetMetadata() plugin.PluginMetadata {
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

func (self *MysqlPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "mysql" {
		services, err := self.SdkClient.GetMysqlServices(cliConnection)
		if err != nil {
			fmt.Fprintf(self.Err, "Unable to retrieve services: %s\n", err)
			self.ExitCode = 1
		}

		if len(services) > 0 {
			fmt.Fprintln(self.Out, "MySQL databases bound to an app:\n")
			for _, service := range (services) {
				fmt.Fprintf(self.Out, "%s\n", service.Name)
			}
		} else {
			fmt.Fprintf(self.Err, "No MySQL databases available. Please bind your database services to " +
				"a started app to make them available to 'cf %s'.\n", args[0])
		}
	}
}

func NewPlugin() *MysqlPlugin {
	return &MysqlPlugin{
		In: os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
		SdkClient: new(SdkCfClient),
	}
}
