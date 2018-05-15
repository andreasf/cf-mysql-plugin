package cfmysql

import "os"

//go:generate counterfeiter . OsWrapper
type OsWrapper interface {
	LookupEnv(key string) (string, bool)
}

func NewOsWrapper() OsWrapper {
	return new(osWrapper)
}

type osWrapper struct {}

func (self *osWrapper) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}