package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/stevenxie/vaingogh/internal/info"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts.",
	Long: `To load completions, run:

. <(` + info.Namespace + ` completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(` + info.Namespace + `completion)
`,
	Run: func(cmd *cobra.Command, args []string) {
		app.GenBashCompletion(os.Stdout)
	},
}
