package cfmysql

import (
	"code.cloudfoundry.org/cli/plugin"
	. "code.cloudfoundry.org/cli/plugin/models"
	"fmt"
)

//go:generate counterfeiter . ApiClient
type ApiClient interface {
	GetStartedApps(cliConnection plugin.CliConnection) ([]GetAppsModel, error)
}

type ApiClientImpl struct {}

func (self *ApiClientImpl) GetStartedApps(cliConnection plugin.CliConnection) ([]GetAppsModel, error) {
	apps, err := cliConnection.GetApps()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve apps: %s", err)
	}

	startedApps := make([]GetAppsModel, 0, len(apps))

	for _, app := range (apps) {
		if app.State == "started" {
			startedApps = append(startedApps, app)
		}
	}

	return startedApps, nil
}