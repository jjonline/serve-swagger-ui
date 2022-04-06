package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/jjonline/serve-swagger-ui/conf"
)

// init version sub-command
func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "show version info",
		Long:  "show version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(conf.Config.Server.Version)
		},
	})
}
