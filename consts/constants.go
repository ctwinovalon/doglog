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

	GreyEsc    = "\033[37m"
	RedEsc     = "\033[91m"
	GreenEsc   = "\033[92m"
	YellowEsc  = "\033[93m"
	BlueEsc    = "\033[94m"
	MagentaEsc = "\033[95m"
	CyanEsc    = "\033[96m"
	WhiteEsc   = "\033[97m"

	ResetEsc = "\033[39;49m"

	DebugEsc = BlueEsc
	ErrorEsc = RedEsc
	InfoEsc  = GreenEsc
	WarnEsc  = YellowEsc

	DebugLevel = "DEBUG"
	ErrorLevel = "ERROR"
	FatalLevel = "FATAL"
	InfoLevel  = "INFO"
	TraceLevel = "TRACE"
	WarnLevel  = "WARN"
)
