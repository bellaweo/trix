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
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", viper.GetString("user"), "matrix username")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	rootCmd.PersistentFlags().StringVarP(&pass, "pass", "p", viper.GetString("pass"), "matrix password")
	viper.BindPFlag("pass", rootCmd.PersistentFlags().Lookup("pass"))
	rootCmd.PersistentFlags().StringVarP(&host, "host", "t", viper.GetString("host"), "matrix hostname")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	rootCmd.PersistentFlags().StringVarP(&room, "room", "r", viper.GetString("room"), "matrix roomid")
	viper.BindPFlag("room", rootCmd.PersistentFlags().Lookup("room"))
}

func vars() {

	var exit = false

	if len(user) == 0 {
		fmt.Println("\nmatrix username required.")
		fmt.Println("set it via the environment variable TRIX_USER or the command line flag --user.")
		exit = true
	}
	if len(pass) == 0 {
		fmt.Println("\nmatrix password required.")
		fmt.Println("set it via the environment variable TRIX_PASS or the command line flag --pass.")
		exit = true
	}
	if len(host) == 0 {
		fmt.Println("\nmatrix host required.")
		fmt.Println("set it via the environment variable TRIX_HOST or the command line flag --host.")
		exit = true
	}
	if len(room) == 0 {
		fmt.Println("\nmatrix room required.")
		fmt.Println("set it via the environment variable TRIX_ROOM or the command line flag --room.")
		exit = true
	}
	if exit {
		fmt.Println("\nrun \"trix help\" for more details")
		os.Exit(1)
	}

}
