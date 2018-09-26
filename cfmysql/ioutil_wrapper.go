package cfmysql

import (
	"io/ioutil"
	"os"
)

//go:generate counterfeiter . IoUtilWrapper
type IoUtilWrapper interface {
	TempFile(dir, pattern string) (f *os.File, err error)
}

func NewIoUtilWrapper() IoUtilWrapper {
	return new(ioUtilWrapper)
}

type ioUtilWrapper struct {}

func (self *ioUtilWrapper) TempFile(dir, pattern string) (f *os.File, err error) {
	return ioutil.TempFile(dir, pattern)
}
