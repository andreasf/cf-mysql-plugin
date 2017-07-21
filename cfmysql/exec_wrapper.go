package cfmysql

import "os/exec"

//go:generate counterfeiter . ExecWrapper
type ExecWrapper interface {
	LookPath(file string) (string, error)
	Run(*exec.Cmd) error
}

type execWrapper struct{}

func NewExecWrapper() ExecWrapper {
	return new(execWrapper)
}

func (self *execWrapper) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (self *execWrapper) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}
