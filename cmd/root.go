package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	user string // matrix username
	pass string // matrix password
	host string // matrix hostname
	room string // matrix roomid

	rootCmd = &cobra.Command{
		Use:   "trix",
		Short: "matrix cli",
		Long:  "a matrix cli for performing one-off tasks.",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(0)
		},
	}
)

// Execute comment
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viper.SetEnvPrefix("TRIX")
	viper.AutomaticEnv()
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", viper.GetString("user"), "matrix username. (or env var TRIX_USER.)")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	rootCmd.PersistentFlags().StringVarP(&pass, "pass", "p", viper.GetString("pass"), "matrix password. (or env var TRIX_PASS.)")
	viper.BindPFlag("pass", rootCmd.PersistentFlags().Lookup("pass"))
	rootCmd.PersistentFlags().StringVarP(&host, "host", "o", viper.GetString("host"), "matrix hostname. (or env var TRIX_HOST.)")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	rootCmd.PersistentFlags().StringVarP(&room, "room", "r", viper.GetString("room"), "matrix roomid. (or env var TRIX_ROOM.)")
	viper.BindPFlag("room", rootCmd.PersistentFlags().Lookup("room"))
}

func vars() (bool, string) {
	var exit = false
	var text string
	if len(user) == 0 {
		text += "\nERROR: matrix username required.\n"
		text += "set it via the environment variable TRIX_USER or the command line flag --user.\n"
		exit = true
	}
	if len(pass) == 0 {
		text += "\nERROR: matrix password required.\n"
		text += "set it via the environment variable TRIX_PASS or the command line flag --pass.\n"
		exit = true
	}
	if len(host) == 0 {
		text += "\nERROR: matrix host required.\n"
		text += "set it via the environment variable TRIX_HOST or the command line flag --host.\n"
		exit = true
	}
	if len(room) == 0 {
		text += "\nERROR: matrix room required.\n"
		text += "set it via the environment variable TRIX_ROOM or the command line flag --room.\n"
		exit = true
	}
	return exit, text
}
