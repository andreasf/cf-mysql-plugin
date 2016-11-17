package cfmysql

import (
	"os/exec"
	"strconv"
	"code.cloudfoundry.org/cli/cf/errors"
	"fmt"
	"os"
)

//go:generate counterfeiter . MysqlRunner
type MysqlRunner interface {
	RunMysql(hostname string, port int, dbName string, username string, password string) error
}

type MysqlClientRunner struct {
	ExecWrapper Exec
}

func (self *MysqlClientRunner) RunMysql(hostname string, port int, dbName string, username string, password string) error {
	path, err := self.ExecWrapper.LookPath("mysql")
	if err != nil {
		return errors.New("'mysql' client not found in PATH'")
	}

	cmd := exec.Command(path, "-u", username, "-p" + password, "-h", hostname, "-P", strconv.Itoa(port), dbName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = self.ExecWrapper.Run(cmd)
	if err != nil {
		return fmt.Errorf("Error running mysql client: %s", err)
	}

	return nil
}

func (self *MysqlClientRunner) MakeMysqlCommand(hostname string, port int, dbName string, username string, password string) *exec.Cmd {
	return exec.Command("mysql", "-u", "username", "-p" + password, "-h", "hostname", "-P", strconv.Itoa(port), dbName)
}

func NewMysqlRunner() MysqlRunner {
	return &MysqlClientRunner{
		ExecWrapper: &ExecWrapper{},
	}
}
