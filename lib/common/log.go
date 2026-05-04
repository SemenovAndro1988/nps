package common

import "log"

// Log is a tiny wrapper used by lib/file to avoid importing
// beego/logs (which would create an import cycle through web).
// External code can override LogPrintf to plug in a different logger.
var LogPrintf = func(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func Log(format string, args ...interface{}) {
	LogPrintf(format, args...)
}
