package command

import (
	"fmt"
	"github.com/jjonline/serve-swagger-ui/app/console"
	"github.com/jjonline/serve-swagger-ui/client/initializer"
	"github.com/jjonline/serve-swagger-ui/conf"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd Cobra-based command line root node definition
var (
	RootCmd = &cobra.Command{
		Use:   "serve-swagger-ui",
		Short: "serve-swagger-ui service manage",
		Long: `
A swagger visual web service that can optionally be authenticated by Google or/and Microsoft account,
configured using command line parameters or a configuration file.
The command line parameter value takes precedence and will override the value of the configuration file`,
		Run: func(cmd *cobra.Command, args []string) {
			console.BootStrap()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// step1、init config
			conf.Init()

			// step2、init global client handle
			initializer.Init()

			return nil
		},
	}
)

func init() {
	// command line args
	RootCmd.PersistentFlags().StringVar(&conf.Cmd.ConfigFile, "config", "", "Specify a TOML configuration file, default conf.toml")
	RootCmd.PersistentFlags().StringVar(&conf.Cmd.Host, "host", "", "Specify the host for the web service, default 0.0.0.0")
	RootCmd.PersistentFlags().IntVar(&conf.Cmd.Port, "port", 0, "Specify the port for the web service, default 9080")
	RootCmd.PersistentFlags().StringVar(&conf.Cmd.SwaggerPath, "path", "", "Specify the swagger JSON file storage path, default ./")
	RootCmd.PersistentFlags().StringVar(&conf.Cmd.LogLevel, "log_level", "", "Specify log level, override config file value：debug|info|warn|error|panic|fatal")
	RootCmd.PersistentFlags().StringVar(&conf.Cmd.LogPath, "log_path", "", "Specify log storage location, override config file value: stderr|stdout|-dir-path-")
	RootCmd.PersistentFlags().BoolVar(&conf.Cmd.OpenBrowser, "open", false, "Automatically open the browser and show the first doc")
}

// Start the app
func Start() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
