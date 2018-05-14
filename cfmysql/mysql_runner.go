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

func NewMysqlRunner(execWrapper ExecWrapper) MysqlRunner {
	return &mysqlRunner{
		execWrapper: execWrapper,
	}
}

type mysqlRunner struct {
	execWrapper ExecWrapper
}

func (self *mysqlRunner) RunMysql(hostname string, port int, dbName string, username string, password string, mysqlArgs ...string) error {
	path, err := self.execWrapper.LookPath("mysql")
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

	err = self.execWrapper.Run(cmd)
	if err != nil {
		return fmt.Errorf("error running mysql client: %s", err)
	}

	return nil
}

func (self *mysqlRunner) RunMysqlDump(hostname string, port int, dbName string, username string, password string, mysqlDumpArgs ...string) error {
	path, err := self.execWrapper.LookPath("mysqldump")
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

	err = self.execWrapper.Run(cmd)
	if err != nil {
		return fmt.Errorf("error running mysqldump: %s", err)
	}

	return nil
}

func (self *mysqlRunner) MakeMysqlCommand(hostname string, port int, dbName string, username string, password string) *exec.Cmd {
	return exec.Command("mysql", "-u", "username", "-p" + password, "-h", "hostname", "-P", strconv.Itoa(port), dbName)
}
