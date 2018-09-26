package cfmysql

import "os"

//go:generate counterfeiter . OsWrapper
type OsWrapper interface {
	LookupEnv(key string) (string, bool)
	Name(file *os.File) string
	Remove(name string) error
	WriteString(file *os.File, s string) (n int, err error)
}

func NewOsWrapper() OsWrapper {
	return new(osWrapper)
}

type osWrapper struct{}

func (self *osWrapper) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (self *osWrapper) Name(file *os.File) string {
	return file.Name()
}

func (self *osWrapper) Remove(name string) error {
	return os.Remove(name)
}

func (self *osWrapper) WriteString(file *os.File, s string) (n int, err error) {
	return file.WriteString(s)
}
