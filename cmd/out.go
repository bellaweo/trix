package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var out = &cobra.Command{
	Use:   "out",
	Short: "send text to the matrix channel",
	Long:  `send string output to the matrix channel`,
	Run: func(cmd *cobra.Command, args []string) {
		vars()
		fmt.Println("blah")
	},
}

func init() {
	rootCmd.AddCommand(out)
}
