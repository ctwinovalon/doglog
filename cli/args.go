package cli

import (
	"doglog/config"
	"doglog/log"
	"doglog/options"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/akamensky/argparse"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// DefaultLimit is the value used when no limit is provided by the user
const DefaultLimit = 300

// DefaultRange is the value used when no range is provided by the user
const DefaultRange = "now-15m"

// DefaultConfigPath is the default location of the configuration path.
const DefaultConfigPath = "~/.doglog"

var defaultIndices = []string{"main"}

// ParseArgs parses the command-line arguments.
// returns: *options which contains both the parsed command-line arguments.
func ParseArgs() *options.Options {
	parser := argparse.NewParser("datadog", "Search and tail logs from Datadog.")

	var defaultConfigPath = expandPath(DefaultConfigPath)

	parser.HelpFunc = customHelp

	service := parser.String("s", "service", &argparse.Options{Required: false, Help: "Special case to search the 'service' message field, e.g., -s send-email is equivalent to -q 'service:send-email'. Merged with the -q query using 'AND' if the -q query is present."})
	query := parser.String("q", "query", &argparse.Options{Required: false, Help: "Query terms to search on (Doglog search syntax). Defaults to '*'."})
	limit := parser.Int("l", "limit", &argparse.Options{Required: false, Help: "The maximum number of messages to request from Datadog. Must be greater then 0", Default: DefaultLimit})
	tail := parser.Flag("t", "tail", &argparse.Options{Required: false, Help: "Whether to tail the output. Requires a relative search."})
	configPath := parser.String("c", "config", &argparse.Options{Required: false, Help: "Path to the config file", Default: defaultConfigPath})
	start := parser.String("", "start", &argparse.Options{Required: false, Help: "Starting date/time to search from. Uses Datadog format", Default: DefaultRange})
	end := parser.String("", "end", &argparse.Options{Required: false, Help: "Ending date/time to search from. Uses Datadog format. Defaults to 'now' if --start is provided but no --end", Default: "now"})
	json := parser.Flag("j", "json", &argparse.Options{Required: false, Help: "Output messages in json format. Shows the modified log message, not the untouched message from Datadog. Useful in understanding the fields available when creating Format templates or for further processing."})
	noColor := parser.Flag("", "no-colors", &argparse.Options{Required: false, Help: "Don't use colors in output."})
	debug := parser.Flag("d", "debug", &argparse.Options{Required: false, Help: "Generate debug output."})
	indexes := parser.StringList("i", "indices", &argparse.Options{Required: false, Help: "The list of indices to search in Datadog. Repeat the parameter to add indices to the list", Default: defaultIndices})
	long := parser.Flag("", "long", &argparse.Options{Required: false, Help: "Generate long output.", Default: false})

	if err := parser.Parse(os.Args); err != nil {
		invalidArgs(parser, err, "")
	}

	if *limit <= 0 {
		var newLimit = DefaultLimit
		limit = &newLimit
	}

	var newQuery string
	if len(*service) > 0 {
		newQuery = "service:" + *service
		if len(*query) > 0 {
			newQuery += " AND " + *query
		}
		query = &newQuery
	}

	if start == nil {
		start = datadog.PtrString("")
	}

	opts := options.Options{
		Service:    *service,
		Query:      *query,
		Limit:      *limit,
		Tail:       *tail,
		ConfigPath: *configPath,
		StartDate:  *start,
		EndDate:    *end,
		Json:       *json,
		Color:      !*noColor && isTty(),
		Debug:      *debug,
		Indexes:    *indexes,
		Long:       *long,
	}

	// Read the configuration file
	conf, err := config.New(opts.ConfigPath)
	if err != nil {
		invalidArgs(parser, err, "")
	}

	opts.ServerConfig = conf

	log.Debug(opts, "Computed query '%s'", opts.Query)

	return &opts
}

// Display the help message when a command-line argument is invalid.
func invalidArgs(parser *argparse.Parser, err error, msg string) {
	if len(msg) > 0 {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n\n", msg, err.Error())
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", msg)
		}
	} else if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", err.Error())
	}
	_, _ = fmt.Fprintf(os.Stderr, customHelp(&parser.Command, nil))
	os.Exit(1)
}

func customHelp(c *argparse.Command, _ interface{}) string {
	buffer := c.Usage(nil)
	return buffer
}

// Expand a leading tilde (~) in a file path into the user's home directory.
func expandPath(configPath string) string {
	var path = configPath
	if strings.HasPrefix(configPath, "~/") {
		usr, _ := user.Current()
		dir := usr.HomeDir

		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}
	return path
}

// Check to see whether we're outputting to a terminal or if we've been redirected to a file
func isTty() bool {
	//_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TIOCGETA)
	//return err == nil
	return true
}
