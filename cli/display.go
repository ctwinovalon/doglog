package cli

import (
	"bytes"
	"doglog/config"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/Masterminds/sprig/v3"
	"strings"
	"text/template"
	"time"
)

const greyEsc = "\033[37m"
const redEsc = "\033[91m"
const greenEsc = "\033[92m"
const yellowEsc = "\033[93m"
const blueEsc = "\033[94m"
const magentaEsc = "\033[95m"
const cyanEsc = "\033[96m"
const whiteEsc = "\033[97m"

const resetEsc = "\033[39;49m"

const debugEsc = blueEsc
const errorEsc = redEsc
const infoEsc = greenEsc
const warnEsc = yellowEsc

const debugLevel = "DEBUG"
const errorLevel = "ERROR"
const fatalLevel = "FATAL"
const infoLevel = "INFO"
const traceLevel = "TRACE"
const warnLevel = "WARN"

// Format a log message into JSON.
func formatJson(msg datadogV2.Log) string {
	var text string

	buf, _ := json.Marshal(msg.AdditionalProperties)
	text = strings.TrimRight(string(buf), "}")
	buf, _ = json.Marshal(msg.GetAttributes().Tags)
	text += ",\"tags\":"
	text += string(buf)
	text += "}"

	text = strings.ReplaceAll(text, "\\\"", "\"")

	return text
}

// Print a single log message
func printMessage(opts *Options, msg *datadogV2.Log) {
	adjustMessage(opts, msg)

	var text string

	jsonField := msg.AdditionalProperties[jsonField]
	if opts.json && jsonField != nil {
		text = jsonField.(string)
	} else {
		for _, f := range opts.serverConfig.Formats() {
			text = tryFormat(*msg, f.Name, f.Format)
			if len(text) > 0 {
				break
			}
		}
	}

	if len(text) > 0 {
		if strings.HasPrefix(text, config.NoFormatDefined) {
			fmt.Println("stop")
		}
		fmt.Println(text)
	} else if jsonField != nil {
		// Last case fallback in case none of the formats (including the default) match
		text = jsonField.(string)
		fmt.Println(text)
	}
}

// Try to apply a format template.
// returns: empty string if the format failed.
func tryFormat(msg datadogV2.Log, tmplName string, tmpl string) string {
	var t = template.Must(template.New(tmplName).Funcs(sprig.TxtFuncMap()).Option("missingkey=error").Parse(tmpl))
	var result bytes.Buffer

	if err := t.Execute(&result, msg.AdditionalProperties); err == nil {
		return result.String()
	}

	return ""
}

// "Cleanup" the log message and add helper fields.
func adjustMessage(opts *Options, msg *datadogV2.Log) {
	isTty := opts.color
	if msg.AdditionalProperties == nil {
		msg.AdditionalProperties = make(map[string]interface{})
	}
	additionalProperties := &msg.AdditionalProperties
	if msg.Attributes.Attributes != nil {
		for k, v := range msg.Attributes.Attributes {
			(*additionalProperties)[k] = v
		}
	}
	for k, v := range msg.AdditionalProperties {
		switch v.(type) {
		case map[string]interface{}:
			for kk, vv := range v.(map[string]interface{}) {
				msg.AdditionalProperties[kk] = vv
			}
			delete(msg.AdditionalProperties, k)
		}
	}
	if msg.Attributes.Status != nil {
		(*additionalProperties)["_Status"] = msg.Attributes.Status
	}
	if msg.Attributes.Service != nil {
		(*additionalProperties)["_Service"] = msg.Attributes.Service
	}
	if msg.Attributes.Host != nil {
		(*additionalProperties)["_Host"] = msg.Attributes.Host
	}
	if msg.Attributes.Timestamp != nil {
		(*additionalProperties)["_Timestamp"] = msg.Attributes.Timestamp.Format(time.RFC3339)
	}

	rpf := (*additionalProperties)[requestPageField]
	if rpf != nil {
		requestPage := rpf.(string)
		if len(requestPage) > 1 && !strings.HasPrefix(requestPage, "/") {
			rpf = "/" + requestPage
		}
	}
	classname, _ := opts.serverConfig.MapField(*msg, "classname")
	if len(classname) > 0 {
		(*additionalProperties)[shortClassnameField] = createShortClassname(classname)
	}

	level := normalizeLevel(opts, *msg)
	(*additionalProperties)[computedLevelField] = level

	constructMessageText(opts, *msg)

	setupColors(isTty, level, *msg)
}

// Set up the colors in the message structure.
func setupColors(isTty bool, level string, msg datadogV2.Log) {
	if isTty {
		computeLevelColor(level, msg)
		// Add color escapes
		msg.AdditionalProperties[blueField] = blueEsc
		msg.AdditionalProperties[redField] = redEsc
		msg.AdditionalProperties[greenField] = greenEsc
		msg.AdditionalProperties[yellowField] = yellowEsc
		msg.AdditionalProperties[greyField] = greyEsc
		msg.AdditionalProperties[whiteField] = whiteEsc
		msg.AdditionalProperties[cyanField] = cyanEsc
		msg.AdditionalProperties[magentaField] = magentaEsc
		msg.AdditionalProperties[resetField] = resetEsc
	} else {
		// Add color escapes
		msg.AdditionalProperties[blueField] = ""
		msg.AdditionalProperties[redField] = ""
		msg.AdditionalProperties[greenField] = ""
		msg.AdditionalProperties[yellowField] = ""
		msg.AdditionalProperties[greyField] = ""
		msg.AdditionalProperties[whiteField] = ""
		msg.AdditionalProperties[cyanField] = ""
		msg.AdditionalProperties[magentaField] = ""
		msg.AdditionalProperties[levelColorField] = ""
		msg.AdditionalProperties[resetField] = ""
	}
}

// Construct the "best" version of the log messages main text. This will look in multiple fields, attempt to
// append multi-line text (stacktraces) onto the message text, etc.
func constructMessageText(opts *Options, msg datadogV2.Log) {
	const nestedException = "; nested exception "
	const newlineNnestedException = ";\nnested exception "

	messageText, _ := opts.serverConfig.MapField(msg, "message")
	originalMessage, _ := opts.serverConfig.MapField(msg, "full_message")
	if len(messageText) == 0 {
		messageText = originalMessage
	}
	if strings.Contains(messageText, nestedException) {
		messageText = strings.Replace(messageText, nestedException, newlineNnestedException, -1)
	}
	if len(originalMessage) > 0 && messageText != originalMessage {
		extraInfo := strings.Split(originalMessage, "\n")
		if len(extraInfo) == 2 {
			messageText = messageText + "\n" + extraInfo[1]
		}
		if len(extraInfo) > 2 {
			messageText = messageText + "\n" + strings.Join(extraInfo[1:len(extraInfo)-1], "\n")
		}
	}
	msg.AdditionalProperties[jsonField] = formatJson(msg)
	if len(messageText) == 0 {
		messageText = msg.AdditionalProperties[jsonField].(string)
	}
	// Replace \" with plain "
	messageText = strings.ReplaceAll(messageText, "\\\"", "\"")
	msg.AdditionalProperties[messageTextField] = messageText
}

// Normalize the "level" of the message.
func normalizeLevel(opts *Options, msg datadogV2.Log) string {
	status := msg.GetAttributes().Status
	level := ""
	if status == nil {
		level, _ = opts.serverConfig.MapField(msg, "level")
	} else {
		level = *status
	}
	level = strings.ToUpper(level)
	if strings.HasPrefix(level, "E") {
		level = errorLevel
	} else if strings.HasPrefix(level, "F") {
		level = fatalLevel
	} else if strings.HasPrefix(level, "I") {
		level = infoLevel
	} else if strings.HasPrefix(level, "W") {
		level = warnLevel
	} else if strings.HasPrefix(level, "D") {
		level = debugLevel
	} else if strings.HasPrefix(level, "T") {
		level = traceLevel
	}
	msg.AdditionalProperties[computedLevelField] = level
	return level
}

// Compute the color that should be used to display the log level in the message output.
func computeLevelColor(level string, msg datadogV2.Log) {
	var levelColor string
	switch level {
	case debugLevel, traceLevel:
		levelColor = debugEsc
	case infoLevel:
		levelColor = infoEsc
	case warnLevel:
		levelColor = warnEsc
	case errorLevel, fatalLevel:
		levelColor = errorEsc
	}
	if len(levelColor) > 0 {
		msg.AdditionalProperties[levelColorField] = levelColor
	} else {
		msg.AdditionalProperties[levelColorField] = ""
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
