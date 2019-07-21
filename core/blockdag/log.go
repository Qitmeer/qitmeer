// Copyright (c) 2017-2018 The qitmeer developers

package blockdag

import (
	l "github.com/HalalChain/qitmeer-lib/log"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log l.Logger

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger l.Logger) {
	log = logger
}
