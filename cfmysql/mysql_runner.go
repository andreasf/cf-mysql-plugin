package cfmysql

import (
	"code.cloudfoundry.org/cli/cf/errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//go:generate counterfeiter . MysqlRunner
type MysqlRunner interface {
	RunMysql(hostname string, port int, dbName string, username string, password string, caCert string, args ...string) error
	RunMysqlDump(hostname string, port int, dbName string, username string, password string, caCert string, args ...string) error
}

func NewMysqlRunner(execWrapper ExecWrapper, ioUtilWrapper IoUtilWrapper, osWrapper OsWrapper) MysqlRunner {
	return &mysqlRunner{
		execWrapper:   execWrapper,
		ioUtilWrapper: ioUtilWrapper,
		osWrapper:     osWrapper,
	}
}

type mysqlRunner struct {
	execWrapper   ExecWrapper
	ioUtilWrapper IoUtilWrapper
	osWrapper     OsWrapper
}

func (self *mysqlRunner) RunMysql(hostname string, port int, dbName string, username string, password string, caCert string, mysqlArgs ...string) error {
	path, err := self.execWrapper.LookPath("mysql")
	if err != nil {
		return errors.New("'mysql' client not found in PATH")
	}

	caCertArgs, caCertPath, err := self.storeCaCert(path, caCert)
	defer self.osWrapper.Remove(caCertPath)
	if err != nil {
		return fmt.Errorf("error preparing TLS arguments: %s", err)
	}

	args := []string{"-u", username, "-p" + password, "-h", hostname, "-P", strconv.Itoa(port)}
	args = append(args, caCertArgs...)
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

func (self *mysqlRunner) RunMysqlDump(hostname string, port int, dbName string, username string, password string, caCert string, mysqlDumpArgs ...string) error {
	path, err := self.execWrapper.LookPath("mysqldump")
	if err != nil {
		return errors.New("'mysqldump' not found in PATH")
	}

	var tableArgs []string
	nonTableArgs := mysqlDumpArgs[0:]

	for i, argument := range mysqlDumpArgs {
		if strings.HasPrefix(argument, "-") {
			break
		}

		tableArgs = append(tableArgs, argument)
		nonTableArgs = mysqlDumpArgs[i+1:]
	}

	caCertArgs, caCertPath, err := self.storeCaCert(path, caCert)
	defer self.osWrapper.Remove(caCertPath)
	if err != nil {
		return fmt.Errorf("error preparing TLS arguments: %s", err)
	}

	args := []string{"-u", username, "-p" + password, "-h", hostname, "-P", strconv.Itoa(port)}
	args = append(args, caCertArgs...)
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

func (self *mysqlRunner) storeCaCert(mysqlPath string, caCert string) ([]string, string, error) {
	if caCert == "" {
		return []string{}, "", nil
	}

	caCertFile, err := self.ioUtilWrapper.TempFile("", "mysql-ca-cert.pem")
	if err != nil {
		return []string{}, "", fmt.Errorf("error creating temp file: %s", err)
	}

	caCertPath := self.osWrapper.Name(caCertFile)

	_, err = self.osWrapper.WriteString(caCertFile, caCert)
	if err != nil {
		return []string{}, "", fmt.Errorf("error writing CA certificate to temp file: %s", err)
	}

	return []string{"--ssl-ca=" + caCertPath}, caCertPath, nil
}
