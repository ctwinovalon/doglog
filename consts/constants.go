package consts

const (
	// Special fields

	RequestPageField = "request_page"

	// Built-in datalog fields

	DatadogStatus    = "__Status"
	DatadogService   = "__Service"
	DatadogHost      = "__Host"
	DatadogTimestamp = "__Timestamp"
	DatadogMessage   = "__Message"

	// Computed fields

	ComputedLevelField          = "__level"
	ComputedMessageField        = "__message"
	ComputedJsonField           = "__json"
	ComputedShortClassnameField = "__short_classname"
	ComputedClassNameField      = "__classname"
	ComputedThreadNameField     = "__threadname"
	ComputedTimestampField      = "__timestamp"

	// Escape codes

	LevelColorField = "_Level_color"
	BlueField       = "_Blue"
	RedField        = "_Red"
	GreenField      = "_Green"
	YellowField     = "_Yellow"
	GreyField       = "_Grey"
	WhiteField      = "_White"
	CyanField       = "_Cyan"
	MagentaField    = "_Magenta"
	ResetField      = "_Reset"
)
