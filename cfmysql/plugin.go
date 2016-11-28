package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"io"
	"fmt"
	"os"
)

type MysqlPlugin struct {
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	ApiClient   ApiClient
	MysqlRunner MysqlRunner
	PortFinder  PortFinder
	exitCode    int
}

func (self *MysqlPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "mysql",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 3,
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
					Usage: "Get a list of available databases:\n   " +
						"cf mysql\n\n   " +
						"Open a mysql client to a database:\n   " +
						"cf mysql <service-name> [mysql args...]",
				},
			},
			{
				Name:     "mysqldump",
				HelpText: "Dump a MySQL database",
				UsageDetails: plugin.Usage{
					Usage: "Get a list of available databases:\n   " +
						"cf mysqldump\n\n   " +
						"Dumping all tables in a database:\n   " +
						"cf mysqldump <service-name> [mysqldump args...]\n\n   " +
						"Dumping specific tables in a database:\n   " +
						"cf mysqldump <service-name> [tables...] [mysqldump args...]",
				},
			},
		},
	}
}

func (self *MysqlPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	command := args[0]

	switch command {
	case "mysql":
		fallthrough

	case "mysqldump":
		if len(args) > 1 {
			dbName := args[1]

			mysqlArgs := []string{}
			if len(args) > 2 {
				mysqlArgs = args[2:]
			}

			self.connectTo(cliConnection, command, dbName, mysqlArgs)
		} else {
			self.showServices(cliConnection, command)
		}
	}
}

func (self *MysqlPlugin) GetExitCode() int {
	return self.exitCode
}

func (self *MysqlPlugin) setErrorExit() {
	self.exitCode = 1
}

func (self *MysqlPlugin) connectTo(cliConnection plugin.CliConnection, command string, dbName string, mysqlArgs []string) {
	services, err := self.ApiClient.GetMysqlServices(cliConnection)
	if err != nil {
		fmt.Fprintf(self.Err, "FAILED\nUnable to retrieve services: %s\n", err)
		self.setErrorExit()
		return
	}

	service, serviceFound := getServiceByName(services, dbName)
	if !serviceFound {
		fmt.Fprintf(self.Err, "FAILED\nService '%s' is not bound to an app, not a MySQL database or does not exist in the " +
			"current space.\n", dbName)
		self.setErrorExit()
		return
	}

	startedApps, err := self.ApiClient.GetStartedApps(cliConnection)
	if err != nil {
		fmt.Fprintf(self.Err, "FAILED\nUnable to retrieve started apps: %s\n", err)
		self.setErrorExit()
		return
	}

	if len(startedApps) == 0 {
		fmt.Fprintf(self.Err, "FAILED\nUnable to connect to '%s': no started apps in current space\n", dbName)
		self.setErrorExit()
		return
	}

	tunnelPort := self.PortFinder.GetPort()
	self.ApiClient.OpenSshTunnel(cliConnection, *service, startedApps[0].Name, tunnelPort)

	err = self.runClient(command, "127.0.0.1", tunnelPort, service.DbName, service.Username, service.Password, mysqlArgs...)
	if err != nil {
		fmt.Fprintf(self.Err, "FAILED\n%s", err)
		self.setErrorExit()
	}
}

func getServiceByName(services []MysqlService, dbName string) (*MysqlService, bool) {
	for _, service := range (services) {
		if service.Name == dbName {
			return &service, true
		}
	}
	return nil, false
}

func (self *MysqlPlugin) runClient(command string, hostname string, port int, dbName string, username string, password string, args ...string) error {
	switch command {
	case "mysql":
		return self.MysqlRunner.RunMysql(hostname, port, dbName, username, password, args...)

	case "mysqldump":
		return self.MysqlRunner.RunMysqlDump(hostname, port, dbName, username, password, args...)
	}

	panic(fmt.Errorf("Command not implemented: %s", command))
}

func (self *MysqlPlugin) showServices(cliConnection plugin.CliConnection, command string) {
	services, err := self.ApiClient.GetMysqlServices(cliConnection)
	if err != nil {
		fmt.Fprintf(self.Err, "Unable to retrieve services: %s\n", err)
		self.setErrorExit()
		return
	}

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
		ApiClient: NewSdkApiClient(),
		PortFinder: new(FreePortFinder),
		MysqlRunner: NewMysqlRunner(),
	}
}
