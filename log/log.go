package log

import (
	"doglog/cli"
	"fmt"
)

func Debug(opts cli.Options, msg string, a ...any) {
	value := fmt.Sprintf(msg, a)
	if opts.Debug {
		fmt.Printf(">>> [DEBUG] %s\n", value)
	}
}

func Info(_ cli.Options, msg string, a ...any) {
	value := fmt.Sprintf(msg, a)
	fmt.Printf(">>> [INFO] %s\n", value)
}

func Error(_ cli.Options, msg string, a ...any) {
	value := fmt.Sprintf(msg, a)
	fmt.Printf(">>> [ERROR] %s\n", value)
}
