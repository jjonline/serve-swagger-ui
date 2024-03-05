package command

import (
	"fmt"
	"github.com/jjonline/serve-swagger-ui/stubs"
	"github.com/spf13/cobra"
)

// init output_conf sub-command
func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "output_conf",
		Short: "output all config values",
		Long:  "output all config values",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("# Copy the following output to create a new Configuration file for .toml suffix")
			fmt.Println(stubs.ConfExample)
		},
	})
}
