package define

// Define some global constants, and default variable values

// ① All default variables in the project are defined here,
//
//	which is convenient for unified management of default values,
//	and the IDE can also quickly track where default values are used.
//
// ② Try to prefix the variable name with Default
var (
	Version         = "0.0.1"            // framework version, use git tag auto replace it when GitHub action auto run
	DefaultSiteName = "serve-swagger-ui" // default site name
	DefaultHost     = "0.0.0.0"          // default host
	DefaultPort     = 9080               // default port
	DefaultPath     = "./"               // default swagger JSON file path，executable binary file sibling directory
	DefaultConfig   = "conf.toml"        // default config file path
	DefaultLogPath  = "stdout"           // default log path
	DefaultLogLevel = "error"            // default log level
)
