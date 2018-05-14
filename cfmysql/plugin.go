package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	"io"
	"fmt"
	"code.cloudfoundry.org/cli/plugin/models"
)

type MysqlPlugin struct {
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	CfService   CfService
	MysqlRunner MysqlRunner
	PortFinder  PortFinder
	exitCode    int
}

func NewMysqlPlugin(conf PluginConf) *MysqlPlugin {
	return &MysqlPlugin{
		In:          conf.In,
		Out:         conf.Out,
		Err:         conf.Err,
		CfService:   conf.CfService,
		PortFinder:  conf.PortFinder,
		MysqlRunner: conf.MysqlRunner,
	}
}

func (self *MysqlPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "mysql",
		Version: plugin.VersionType{
			Major: 2,
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
					Usage: "Open a mysql client to a database:\n   " +
						"cf mysql <service-name> [mysql args...]",
				},
			},
			{
				Name:     "mysqldump",
				HelpText: "Dump a MySQL database",
				UsageDetails: plugin.Usage{
					Usage: "Dump all tables in a database:\n   " +
						"cf mysqldump <service-name> [mysqldump args...]\n   " +
						"Dump specific tables in a database:\n   " +
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

			var mysqlArgs []string
			if len(args) > 2 {
				mysqlArgs = args[2:]
			}

			self.connectTo(cliConnection, command, dbName, mysqlArgs)
		} else {
			fmt.Fprint(self.Err, self.FormatUsage())
			self.setErrorExit()
		}

	default:
		// we don't handle "uninstall"
	}
}

func (self *MysqlPlugin) FormatUsage() string {
	var usage string
	for i, command := range self.GetMetadata().Commands {
		if i > 0 {
			usage += "\n\n"
		}

		usage += fmt.Sprintf(
			"cf %s - %s\n\nUSAGE:\n   %s\n",
			command.Name,
			command.HelpText,
			command.UsageDetails.Usage,
		)
	}

	return usage
}

func (self *MysqlPlugin) GetExitCode() int {
	return self.exitCode
}

func (self *MysqlPlugin) setErrorExit() {
	self.exitCode = 1
}

type StartedAppsResult struct {
	Apps []plugin_models.GetAppsModel
	Err  error
}

func (self *MysqlPlugin) connectTo(cliConnection plugin.CliConnection, command string, dbName string, mysqlArgs []string) {
	appsChan := make(chan StartedAppsResult, 0)
	go func() {
		startedApps, err := self.CfService.GetStartedApps(cliConnection)
		appsChan <- StartedAppsResult{Apps: startedApps, Err: err}
	}()

	service, err := self.CfService.GetService(cliConnection, dbName)
	if err != nil {
		fmt.Fprintf(self.Err, "FAILED\nUnable to retrieve service credentials: %s\n", err)
		self.setErrorExit()
		return
	}

	appsResult := <-appsChan
	if appsResult.Err != nil {
		fmt.Fprintf(self.Err, "FAILED\nUnable to retrieve started apps: %s\n", appsResult.Err)
		self.setErrorExit()
		return
	}

	if len(appsResult.Apps) == 0 {
		fmt.Fprintf(self.Err, "FAILED\nUnable to connect to '%s': no started apps in current space\n", dbName)
		self.setErrorExit()
		return
	}

	tunnelPort := self.PortFinder.GetPort()
	self.CfService.OpenSshTunnel(cliConnection, service, appsResult.Apps, tunnelPort)

	err = self.runClient(command, "127.0.0.1", tunnelPort, service.DbName, service.Username, service.Password, mysqlArgs...)
	if err != nil {
		fmt.Fprintf(self.Err, "FAILED\n%s", err)
		self.setErrorExit()
	}
}

func (self *MysqlPlugin) runClient(command string, hostname string, port int, dbName string, username string, password string, args ...string) error {
	switch command {
	case "mysql":
		return self.MysqlRunner.RunMysql(hostname, port, dbName, username, password, args...)

	case "mysqldump":
		return self.MysqlRunner.RunMysqlDump(hostname, port, dbName, username, password, args...)
	}

	panic(fmt.Errorf("command not implemented: %s", command))
}

type PluginConf struct {
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	CfService   CfService
	MysqlRunner MysqlRunner
	PortFinder  PortFinder
}
