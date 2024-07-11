package config

import (
	"doglog/consts"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strings"
)

const NoFormatDefined = "No Formats Defined>>"

const formatsSection string = "formats.short"    // [formats]
const longFormatsSection string = "formats.long" // [formats.long]
const serverSection string = "server"            // [server]
const fieldSection string = "fields"             // [fields]
const apiKey = "api-key"
const applicationKey = "application-key"

// FormatDefinition stores a single format line.
type FormatDefinition struct {
	Name   string
	Format string
}

// IniFile is a wrapper around the INI file reader
type IniFile struct {
	ini *ini.File
	// Stores formats so we don't keep re-reading them
	storedFormats []FormatDefinition
	// Stores field mappings so we don't keep re-reading them
	storedFields map[string][]string
}

// New creates a new INI file reader and wraps it.
func New(configPath string) (*IniFile, error) {
	if f, err := readConfig(configPath); err == nil {
		return &IniFile{ini: f}, nil
	} else {
		return nil, err
	}
}

func (c *IniFile) AllFields() map[string][]string {
	return c.storedFields
}

// ApiKey gets the API key from the config file. Defaults to an empty string.
func (c *IniFile) ApiKey() string {
	server := c.ini.Section(serverSection)
	return server.Key(apiKey).MustString("")
}

// ApplicationKey gets the application key from the config file. Defaults to an empty string.
func (c *IniFile) ApplicationKey() string {
	server := c.ini.Section(serverSection)
	return server.Key(applicationKey).MustString("")
}

// Formats gets the log messages formats from the config file. Adds a final default format case so the user knows that
// no formats were applied successfully.
func (c *IniFile) Formats(useLong bool) []FormatDefinition {
	var formats []FormatDefinition
	if c.storedFormats == nil {
		sectionName := formatsSection
		if useLong {
			sectionName = longFormatsSection
		}
		for _, f := range c.ini.Section(sectionName).Keys() {
			formats = append(formats, FormatDefinition{Name: f.Name(), Format: f.Value()})
		}
		formats = append(formats, FormatDefinition{Name: "_default", Format: NoFormatDefined + " {{." + consts.ComputedJsonField + "}}"})
		c.storedFormats = formats
	}

	return c.storedFormats
}

// Fields gets the field mappings from the config file. These will be merged with the defaults.
func (c *IniFile) Fields() map[string][]string {
	if c.storedFields == nil {
		c.storedFields = make(map[string][]string)
		c.storedFields[consts.ComputedLevelField] =
			[]string{consts.DatadogStatus, "level", "status", "loglevel", "log_status", "LogLevel", "severity"}
		c.storedFields[consts.ComputedMessageField] =
			[]string{consts.DatadogMessage, "message", "msg", "textPayload", "Message"}
		c.storedFields[consts.ComputedClassNameField] =
			[]string{"classname", "logger_name", "LoggerName", "component", "name"}
		c.storedFields[consts.ComputedThreadNameField] =
			[]string{"threadname", "thread_name"}
		c.storedFields[consts.ComputedTimestampField] =
			[]string{consts.DatadogTimestamp, "timestamp"}

		for _, f := range c.ini.Section(fieldSection).Keys() {
			name := f.Name()
			value := f.Value()
			fieldList := strings.Split(value, ",")
			for i := range fieldList {
				fieldList[i] = strings.TrimSpace(fieldList[i])
			}
			c.storedFields[name] = fieldList
		}
	}

	return c.storedFields
}

func (c *IniFile) AggregateFields(msg datadogV2.Log) {
	if msg.AdditionalProperties != nil {
		fieldMappings := c.Fields()
		for k := range fieldMappings {
			if value, ok := c.MapField(msg, k); ok {
				msg.AdditionalProperties[k] = value
			}
		}
	}
}

// MapField Pull a field from the 'fields' map, using field mappings as available
func (c *IniFile) MapField(msg datadogV2.Log, field string) (string, bool) {
	fieldMappings := c.Fields()
	fieldList, ok := fieldMappings[field]
	if !ok {
		fieldList = make([]string, 1)
		fieldList[0] = field
	}

	for _, f := range fieldList {
		if value, ok := msg.AdditionalProperties[f]; ok {
			switch value.(type) {
			case string:
				return value.(string), true
			case *string:
				return *(value.(*string)), true
			}
			return value.(string), true
		}
	}
	return "", false
}

// Reads the configuration file. The configuration is stored in a INI style file.
func readConfig(configPath string) (*ini.File, error) {
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("configuration file not found at %s - %s", configPath, err)
	}

	if _, err2 := os.Stat(configPath); err2 != nil {
		return nil, fmt.Errorf("configuration file not found or not readable at %s - %s", configPath, err2)
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("configuration file cannot be parsed at %s - %s", configPath, err)
	}

	return cfg, nil
}
