package cfmysql

import "os/exec"

//go:generate counterfeiter . ExecWrapper
type ExecWrapper interface {
	CombinedOutput(*exec.Cmd) ([]byte, error)
	LookPath(file string) (string, error)
	Run(*exec.Cmd) error
}

func NewExecWrapper() ExecWrapper {
	return new(execWrapper)
}

type execWrapper struct{}

func (self *execWrapper) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (self *execWrapper) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (self *execWrapper) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}
