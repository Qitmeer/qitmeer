// Copyright (c) 2017-2018 The qitmeer developers

package blkmgr

import (
	l "github.com/Qitmeer/qng-core/log"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log l.Logger

// The default amount of logging is none.
func init() {
	UseLogger(l.New(l.Ctx{"module": "blkmanager"}))
}

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger l.Logger) {
	log = logger
}

// LogClosure is a closure that can be printed with %v to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type logClosure func() string

func (c logClosure) String() string {
	return c()
}

func newLogClosure(c func() string) logClosure {
	return logClosure(c)
}
