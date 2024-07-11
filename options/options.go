package options

import "doglog/config"

// Options Structure stores the command-line options and values.
type Options struct {
	Service      string
	Query        string
	Limit        int
	DoTail       bool
	ConfigPath   string
	TimeRange    int
	StartDate    string
	EndDate      string
	OutputJson   bool
	ServerConfig *config.IniFile
	UseColor     bool
	PrintDebug   bool
	Indexes      []string
	UseLong      bool
	Version      bool
}
