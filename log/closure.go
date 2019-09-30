package log

// LogClosure is a closure that can be printed with %v to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type LogClosure func() string

func (c LogClosure) String() string {
	return c()
}

func newLogClosure(c func() string) LogClosure {
	return LogClosure(c)
}
