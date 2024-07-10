package options

import "doglog/config"

// Options structure stores the command-line Options and values.
type Options struct {
	Service      string
	Query        string
	Limit        int
	Tail         bool
	ConfigPath   string
	TimeRange    int
	StartDate    string
	EndDate      string
	Json         bool
	ServerConfig *config.IniFile
	Color        bool
	Debug        bool
	Indexes      []string
	Long         bool
}
