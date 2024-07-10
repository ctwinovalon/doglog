package log

import (
	"doglog/options"
	"fmt"
)

func Debug(opts options.Options, msg string, a ...any) {
	if opts.Debug {
		value := fmt.Sprintf(msg, a)
		fmt.Printf(">>> [DEBUG] %s\n", value)
	}
}

func Info(opts options.Options, msg string, a ...any) {
	if opts.Debug {
		value := fmt.Sprintf(msg, a)
		fmt.Printf(">>> [INFO] %s\n", value)
	}
}

func Error(opts options.Options, msg string, a ...any) {
	if opts.Debug {
		value := fmt.Sprintf(msg, a)
		fmt.Printf(">>> [ERROR] %s\n", value)
	}
}
