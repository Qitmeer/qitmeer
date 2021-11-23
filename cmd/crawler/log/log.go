/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package log

import (
	l "github.com/Qitmeer/qng-core/log"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var Log l.Logger

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger l.Logger) {
	Log = logger
}

// The default amount of logging is none.
func init() {
	UseLogger(l.New(l.Ctx{"module": "CRAWLER"}))
}
