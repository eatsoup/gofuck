package conf

// Settings holds user/runtime preferences that rules can read.
// This is a reasonable subset of thefuck's conf.py.
type Settings struct {
	Rules                       []string
	ExcludeRules                []string
	RequireConfirmation         bool
	WaitCommand                 int
	NoColors                    bool
	Debug                       bool
	HistoryLimit                int
	AlterHistory                bool
	WaitSlowCommand             int
	SlowCommands                []string
	NumCloseMatches             int
	EnvAllowList                []string
	Priority                    map[string]int
	Repeat                      bool
	InstantMode                 bool
	UserDir                     string
	ExcludedSearchPathPrefixes  []string
}

const ALL_ENABLED = "DEFAULT_RULES"
const DEFAULT_PRIORITY = 1000

// Default returns a fresh Settings with thefuck defaults.
func Default() *Settings {
	return &Settings{
		Rules:                      []string{ALL_ENABLED},
		ExcludeRules:               []string{},
		RequireConfirmation:        true,
		WaitCommand:                3,
		NoColors:                   false,
		Debug:                      false,
		HistoryLimit:               0,
		AlterHistory:               true,
		WaitSlowCommand:            15,
		SlowCommands:               []string{"lein", "react-native", "gradle", "./gradlew", "vagrant"},
		NumCloseMatches:            3,
		Priority:                   map[string]int{},
		Repeat:                     false,
		InstantMode:                false,
		ExcludedSearchPathPrefixes: []string{},
	}
}

// Current is the active settings object; rules may read from it.
var Current = Default()

// Reset resets Current to defaults. Mainly useful in tests.
func Reset() { Current = Default() }
