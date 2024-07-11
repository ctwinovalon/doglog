package main

import (
	"doglog/cli"
	"fmt"
	"github.com/briandowns/spinner"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// The appVersion is filled in during the build
//
//goland:noinspection GoUnusedGlobalVariable
var appVersion = "v0.0.0"

// The gitHash is filled in during the build
var gitHash = ""

// Create a new terminal spinner.
func setupSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.UpdateCharSet(spinner.CharSets[21]) // box of dots
	s.Writer = os.Stderr
	s.HideCursor = true
	_ = s.Color("red", "bold")
	return s
}

// This channel is purely for the handling of signals.
func makeSignalsChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGKILL, // "always fatal", "SIGKILL and SIGSTOP may not be caught by a program"
		syscall.SIGHUP,  // "terminal is disconnected"
	)
	return c
}

func main() {
	version := fmt.Sprintf("%v (%v)", appVersion, gitHash)
	opts := cli.ParseArgs(version)

	if opts.DoTail {
		var delay = cli.MinDelay

		s := setupSpinner()
		s.Start()

		exitChan := makeSignalsChannel()

		// Handle exit signals - only needed when tailing
		go func() {
			for range exitChan {
				s.Stop()
				os.Exit(0)
			}
		}()

		//noinspection GoInfiniteFor
		for {
			found := cli.CommandListMessages(opts, s)
			delay = cli.DelayForSeconds(delay, found)
		}
	} else {
		_ = cli.CommandListMessages(opts, nil)
	}
}
