/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:log.go
 * Date:7/7/20 3:06 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package discover

import (
	l "github.com/Qitmeer/qitmeer/log"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log l.Logger

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger l.Logger) {
	log = logger
}

// The default amount of logging is none.
func init() {
	UseLogger(l.New(l.Ctx{"module": "DISC"}))
}
