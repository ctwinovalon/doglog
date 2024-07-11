package cli

import (
	"bytes"
	"doglog/config"
	"doglog/consts"
	"doglog/log"
	"doglog/options"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/Masterminds/sprig/v3"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Format a log message into JSON.
func formatJson(msg datadogV2.Log) (text string) {
	buf, _ := json.Marshal(msg.AdditionalProperties)
	text = strings.TrimRight(string(buf), "}")
	buf, _ = json.Marshal(msg.GetAttributes().Tags)
	text += ",\"tags\":"
	text += string(buf)
	text += "}"

	text = strings.ReplaceAll(text, "\\\"", "\"")

	return
}

// Print a single log message to stdout.
func printMessage(opts *options.Options, msg *datadogV2.Log) {
	adjustMap(opts, msg)

	var text string

	jsonField := msg.AdditionalProperties[consts.ComputedJsonField]
	if opts.Json && jsonField != nil {
		text = jsonField.(string)
	} else {
		for _, f := range opts.ServerConfig.Formats(opts.Long) {
			text = tryFormat(opts, *msg, f.Name, f.Format)
			if len(text) > 0 {
				break
			}
		}
	}

	if len(text) > 0 {
		if strings.HasPrefix(text, config.NoFormatDefined) {
			fmt.Println("")
		}
		fmt.Println(text)
		if strings.HasPrefix(text, config.NoFormatDefined) {
			fmt.Println("")
		}
	} else if jsonField != nil {
		// Last case fallback in case none of the formats (including the default) match
		text = jsonField.(string)
		fmt.Println(text)
	}
}

// Try to apply a format template and return an empty string if the format failed.
func tryFormat(opts *options.Options, msg datadogV2.Log, tmplName string, tmpl string) string {
	var t = template.Must(template.New(tmplName).Funcs(sprig.TxtFuncMap()).Option("missingkey=error").Parse(tmpl))
	var result bytes.Buffer

	err := t.Execute(&result, msg.AdditionalProperties)
	if err == nil {
		log.Info(*opts, "Applied template '%s' successfully", tmplName)
		return result.String()
	}
	log.Debug(*opts, "failed to apply template '%s': %v", tmplName, err)

	return ""
}

// Collapse a tree of maps into a single top-level map.
func flatten(src map[string]interface{}, dest map[string]interface{}) {
	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			flatten(child, dest)
		case []interface{}:
			for i := 0; i < len(child); i++ {
				dest[k+"."+strconv.Itoa(i)] = child[i]
			}
		default:
			dest[k] = v
		}
	}
}

// Normalize the log message and add helper fields.
func adjustMap(opts *options.Options, msg *datadogV2.Log) {
	isTty := opts.Color
	if msg.AdditionalProperties == nil {
		msg.AdditionalProperties = make(map[string]interface{})
	}
	additionalProperties := &msg.AdditionalProperties
	if msg.Attributes.Attributes != nil {
		flatten(msg.Attributes.Attributes, msg.AdditionalProperties)
	}
	if msg.Attributes.Status != nil {
		(*additionalProperties)[consts.DatadogStatus] = msg.Attributes.Status
	}
	if msg.Attributes.Service != nil {
		(*additionalProperties)[consts.DatadogService] = msg.Attributes.Service
	}
	if msg.Attributes.Host != nil {
		(*additionalProperties)[consts.DatadogHost] = msg.Attributes.Host
	}
	if msg.Attributes.Timestamp != nil {
		(*additionalProperties)[consts.DatadogTimestamp] = msg.Attributes.Timestamp.Format(time.RFC3339)
	}
	if msg.Attributes.Message != nil {
		(*additionalProperties)[consts.DatadogMessage] = msg.Attributes.Message
	}

	opts.ServerConfig.AggregateFields(*msg)

	if rpf := (*additionalProperties)[consts.RequestPageField]; rpf != nil {
		requestPage := rpf.(string)
		if len(requestPage) > 1 && !strings.HasPrefix(requestPage, "/") {
			rpf = "/" + requestPage
		}
	}
	classname := getField(msg.AdditionalProperties, consts.ComputedClassNameField)
	if len(classname) > 0 {
		(*additionalProperties)[consts.ComputedShortClassnameField] = createShortClassname(classname)
	}

	level := normalizeLevel(*msg)
	(*additionalProperties)[consts.ComputedLevelField] = level

	constructMessageText(*msg)

	msg.AdditionalProperties[consts.ComputedJsonField] = formatJson(*msg)

	setupColors(isTty, level, *msg)
}

// Extract a named entry from a map, returning an empty string if not found.
func getField(props map[string]interface{}, field string) string {
	if value, ok := props[field]; ok {
		switch v := value.(type) {
		case string:
			return v
		case *string:
			return *v
		default:
			return ""
		}
	} else {
		return ""
	}
}

// Set up the colors in the message structure.
func setupColors(isTty bool, level string, msg datadogV2.Log) {
	if isTty {
		computeLevelColor(level, msg)
		// Add color escapes
		msg.AdditionalProperties[consts.BlueField] = consts.BlueEsc
		msg.AdditionalProperties[consts.RedField] = consts.RedEsc
		msg.AdditionalProperties[consts.GreenField] = consts.GreenEsc
		msg.AdditionalProperties[consts.YellowField] = consts.YellowEsc
		msg.AdditionalProperties[consts.GreyField] = consts.GreyEsc
		msg.AdditionalProperties[consts.WhiteField] = consts.WhiteEsc
		msg.AdditionalProperties[consts.CyanField] = consts.CyanEsc
		msg.AdditionalProperties[consts.MagentaField] = consts.MagentaEsc
		msg.AdditionalProperties[consts.ResetField] = consts.ResetEsc
	} else {
		// Add color escapes
		msg.AdditionalProperties[consts.BlueField] = ""
		msg.AdditionalProperties[consts.RedField] = ""
		msg.AdditionalProperties[consts.GreenField] = ""
		msg.AdditionalProperties[consts.YellowField] = ""
		msg.AdditionalProperties[consts.GreyField] = ""
		msg.AdditionalProperties[consts.WhiteField] = ""
		msg.AdditionalProperties[consts.CyanField] = ""
		msg.AdditionalProperties[consts.MagentaField] = ""
		msg.AdditionalProperties[consts.LevelColorField] = ""
		msg.AdditionalProperties[consts.ResetField] = ""
	}
}

// Construct the "best" version of the log messages main text. This will look in multiple fields, attempt to
// append multi-line text (stacktraces) onto the message text, etc.
func constructMessageText(msg datadogV2.Log) {
	messageText := getField(msg.AdditionalProperties, consts.ComputedMessageField)
	// Replace \" with plain "
	messageText = strings.ReplaceAll(messageText, "\\\"", "\"")
	msg.AdditionalProperties[consts.ComputedMessageField] = messageText
}

// Normalize the "level" of the message.
func normalizeLevel(msg datadogV2.Log) (level string) {
	if status := msg.GetAttributes().Status; status == nil {
		level = getField(msg.AdditionalProperties, consts.ComputedLevelField)
	} else {
		level = *status
	}
	level = strings.ToUpper(level)
	if strings.HasPrefix(level, "E") {
		level = consts.ErrorLevel
	} else if strings.HasPrefix(level, "F") {
		level = consts.FatalLevel
	} else if strings.HasPrefix(level, "I") {
		level = consts.InfoLevel
	} else if strings.HasPrefix(level, "W") {
		level = consts.WarnLevel
	} else if strings.HasPrefix(level, "D") {
		level = consts.DebugLevel
	} else if strings.HasPrefix(level, "T") {
		level = consts.TraceLevel
	}
	msg.AdditionalProperties[consts.ComputedLevelField] = level
	return
}

// Compute the color that should be used to display the log level in the message output.
func computeLevelColor(level string, msg datadogV2.Log) {
	var levelColor string
	switch level {
	case consts.DebugLevel, consts.TraceLevel:
		levelColor = consts.DebugEsc
	case consts.InfoLevel:
		levelColor = consts.InfoEsc
	case consts.WarnLevel:
		levelColor = consts.WarnEsc
	case consts.ErrorLevel, consts.FatalLevel:
		levelColor = consts.ErrorEsc
	}
	if len(levelColor) > 0 {
		msg.AdditionalProperties[consts.LevelColorField] = levelColor
	} else {
		msg.AdditionalProperties[consts.LevelColorField] = ""
	}
}

// Create a shortened version of the Java classname.
func createShortClassname(classname string) string {
	parts := strings.Split(classname, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return classname
}
