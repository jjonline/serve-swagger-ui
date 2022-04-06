package initializer

import "github.com/jjonline/serve-swagger-ui/client"

// region Global handle initialization related

// Init Initialize
//go:noinline
func Init() {
	client.Logger = iniLogger()            // Initialize the logger, which needs to be executed first
	client.MemoryCache = initMemoryCache() // Initialize the memory cache
	client.Guzzle = initGuzzle()           // init common http client
}

// endregion

// region Global handle initialization related after hot reload

// Reload hot reload initialization call again
//  - When the configuration changes, there is no need to restart the process to trigger, monitor the call
func Reload() {

}

// endregion
