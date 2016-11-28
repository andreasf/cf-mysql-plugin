package cfmysql

import (
	"os/exec"
	"strconv"
	"code.cloudfoundry.org/cli/cf/errors"
	"fmt"
	"os"
	"strings"
)

//go:generate counterfeiter . MysqlRunner
type MysqlRunner interface {
	RunMysql(hostname string, port int, dbName string, username string, password string, args ...string) error
	RunMysqlDump(hostname string, port int, dbName string, username string, password string, args ...string) error
}

type MysqlClientRunner struct {
	ExecWrapper Exec
}

func (self *MysqlClientRunner) RunMysql(hostname string, port int, dbName string, username string, password string, mysqlArgs ...string) error {
	path, err := self.ExecWrapper.LookPath("mysql")
	if err != nil {
		return errors.New("'mysql' client not found in PATH")
	}

	args := []string{"-u", username, "-p" + password, "-h", hostname, "-P", strconv.Itoa(port)}
	args = append(args, mysqlArgs...)
	args = append(args, dbName)

	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = self.ExecWrapper.Run(cmd)
	if err != nil {
		return fmt.Errorf("Error running mysql client: %s", err)
	}

	return nil
}

func (self *MysqlClientRunner) RunMysqlDump(hostname string, port int, dbName string, username string, password string, mysqlDumpArgs ...string) error {
	path, err := self.ExecWrapper.LookPath("mysqldump")
	if err != nil {
		return errors.New("'mysqldump' not found in PATH")
	}

	tableArgs := []string{}
	nonTableArgs := mysqlDumpArgs[0:]

	for i, argument := range (mysqlDumpArgs) {
		if strings.HasPrefix(argument, "-") {
			break
		}

		tableArgs = append(tableArgs, argument)
		nonTableArgs = mysqlDumpArgs[i + 1:]
	}

	args := []string{"-u", username, "-p" + password, "-h", hostname, "-P", strconv.Itoa(port)}
	args = append(args, nonTableArgs...)
	args = append(args, dbName)
	args = append(args, tableArgs...)

	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = self.ExecWrapper.Run(cmd)
	if err != nil {
		return fmt.Errorf("Error running mysqldump: %s", err)
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
