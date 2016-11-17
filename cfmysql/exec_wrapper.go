package cfmysql

import "os/exec"

//go:generate counterfeiter . Exec
type Exec interface {
	LookPath(file string) (string, error)
	Run(*exec.Cmd) error
}

type ExecWrapper struct{}

func (self *ExecWrapper) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (self *ExecWrapper) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}
