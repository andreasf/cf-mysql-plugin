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
	ApiClient ApiClient
	exitCode  int
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
	command := args[0]
	switch command {
	case "mysql":
		services, err := self.ApiClient.GetMysqlServices(cliConnection)
		if err != nil {
			fmt.Fprintf(self.Err, "Unable to retrieve services: %s\n", err)
			self.setErrorExit()
			return
		}

		if len(args) > 1 {
			dbName := args[1]
			self.connectTo(services, command, dbName)
		} else {
			self.showServices(services, command)
		}
	}

}

func (self *MysqlPlugin) GetExitCode() int {
	return self.exitCode
}

func (self *MysqlPlugin) setErrorExit() {
	self.exitCode = 1
}

func (self *MysqlPlugin) connectTo(services []MysqlService, command string, dbName string) {
	fmt.Fprintf(self.Err, "Service '%s' is not bound to an app, not a MySQL database or does not exist in the " +
		"current space.\n", dbName)
	self.setErrorExit()
}

func (self *MysqlPlugin) showServices(services []MysqlService, command string) {
	if len(services) > 0 {
		fmt.Fprintln(self.Out, "MySQL databases bound to an app:\n")
		for _, service := range (services) {
			fmt.Fprintf(self.Out, "%s\n", service.Name)
		}
	} else {
		fmt.Fprintf(self.Err, "No MySQL databases available. Please bind your database services to " +
			"a started app to make them available to 'cf %s'.\n", command)
	}
}

func NewPlugin() *MysqlPlugin {
	return &MysqlPlugin{
		In: os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
		ApiClient: new(SdkApiClient),
	}
}
