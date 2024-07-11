package log

import (
	"doglog/options"
	"fmt"
)

// Debug Output a log message at debug level. This is only used when outputting
// debug/verbose output.
func Debug(opts options.Options, msg string, a ...any) {
	if opts.Debug {
		value := fmt.Sprintf(msg, a)
		fmt.Printf(">>> [DEBUG] %s\n", value)
	}
}

// Info Output a log message at info level. This is only used when outputting
// debug/verbose output.
func Info(opts options.Options, msg string, a ...any) {
	if opts.Debug {
		value := fmt.Sprintf(msg, a)
		fmt.Printf(">>> [INFO] %s\n", value)
	}
}

// Error Output a log message at error level. This is only used when outputting
// debug/verbose output.
func Error(opts options.Options, msg string, a ...any) {
	if opts.Debug {
		value := fmt.Sprintf(msg, a)
		fmt.Printf(">>> [ERROR] %s\n", value)
	}
}
