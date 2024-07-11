package cli

import (
	"doglog/config"
	"doglog/log"
	"doglog/options"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/akamensky/argparse"
	"golang.org/x/term"
	"os"
	"path/filepath"
)

// DefaultLimit is the value used when no limit is provided by the user
const DefaultLimit = 300

// DefaultRange is the value used when no range is provided by the user
const DefaultRange = "now-15m"

// DefaultConfigName is the default location of the configuration path.
const DefaultConfigName = ".doglog"

var defaultIndices = []string{"main"}

var AppVersion = ""

// ParseArgs parses the command-line arguments and returns the *options
// which contain the parsed command-line arguments.
func ParseArgs(appVersion string) *options.Options {
	AppVersion = appVersion
	opts := initializeArgumentParser()

	// Display the application version. Put this here in case there's an error in
	// the succeeding code
	if opts.Version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	return &opts
}

// Set up the argument parser and return the options selected
func initializeArgumentParser() options.Options {
	parser := argparse.NewParser("datadog", "Search and tail logs from Datadog.")
	defaultConfigPath := defaultConfigFile()

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
	version := parser.Flag("v", "version", &argparse.Options{Required: false, Help: "Display the application version and exit."})

	if err := parser.Parse(os.Args); err != nil {
		invalidArgs(parser, err, "")
	}

	if start == nil {
		start = datadog.PtrString("")
	}

	opts := options.Options{
		Service:    *service,
		Query:      *query,
		Limit:      *limit,
		DoTail:     *tail,
		ConfigPath: *configPath,
		StartDate:  *start,
		EndDate:    *end,
		OutputJson: *json,
		UseColor:   !*noColor && isTty(),
		PrintDebug: *debug,
		Indexes:    *indexes,
		UseLong:    *long,
		Version:    *version,
	}

	if opts.Limit <= 0 {
		var newLimit = DefaultLimit
		opts.Limit = newLimit
	}

	opts.ServerConfig = loadConfigFile(opts, parser, &opts.ConfigPath)

	opts.Query = constructQuery(opts.Service, opts.Query)
	log.Debug(opts, "Computed query '%s'", opts.Query)

	return opts
}

// Add 'service:' to the query
func constructQuery(service string, query string) string {
	var newQuery string
	if len(service) > 0 {
		newQuery = "service:" + service
		if len(query) > 0 {
			newQuery += " AND " + query
		}
		query = newQuery
	}
	return query
}

// Load the configuration
func loadConfigFile(opts options.Options, parser *argparse.Parser, configPath *string) *config.IniFile {
	testPath, err := filepath.Abs(*configPath)
	if err != nil {
		invalidArgs(parser, err, "Config file path is invalid")
	}
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		invalidArgs(parser, err, "Config file does not exist")
	}
	// Read the configuration file
	conf, err := config.New(opts.ConfigPath)
	if err != nil {
		invalidArgs(parser, err, "")
	}

	return conf
}

// Determine the default configuration file location
func defaultConfigFile() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Printf("Can't determine user's home directory - %s\n", err)
		os.Exit(1)
	}
	defaultConfigPath, err := filepath.Abs(filepath.Join(dirname, DefaultConfigName))
	if err != nil {
		_, _ = fmt.Printf("Can't determine default config path - %s\n", err)
		os.Exit(1)
	}
	return defaultConfigPath
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

// Generate the help usage text
func customHelp(c *argparse.Command, _ interface{}) (buffer string) {
	buffer = fmt.Sprintf("Version: %s\n", AppVersion)
	buffer += c.Usage(nil)
	return
}

// Check to see whether we're outputting to a terminal or if we've been redirected to a file
func isTty() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
