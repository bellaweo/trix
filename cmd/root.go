package cmd

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type rootVars struct {
	User string // matrix username
	Pass string // matrix password
	Host string // matrix hostname
	Room string // matrix roomid or room alias
}

var (
	root rootVars

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
	rootCmd.PersistentFlags().StringVarP(&root.User, "user", "u", viper.GetString("user"), "matrix username. (env var TRIX_USER.)")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	rootCmd.PersistentFlags().StringVarP(&root.Pass, "pass", "p", viper.GetString("pass"), "matrix password. (env var TRIX_PASS.)")
	viper.BindPFlag("pass", rootCmd.PersistentFlags().Lookup("pass"))
	rootCmd.PersistentFlags().StringVarP(&root.Host, "host", "o", viper.GetString("host"), "matrix hostname. (env var TRIX_HOST.)")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	rootCmd.PersistentFlags().StringVarP(&root.Room, "room", "r", viper.GetString("room"), "matrix roomid or alias. (env var TRIX_ROOM.)")
	viper.BindPFlag("room", rootCmd.PersistentFlags().Lookup("room"))
}

func validateHost(host string) string {
	var text string
	_, err := url.ParseRequestURI(host)
	if err != nil {
		text = fmt.Sprintf("\nERROR: matrix flag host format not valid. %v.\n", err)
		return text
	}
	u, err := url.Parse(host)
	if err != nil || u.Scheme == "" || u.Host == "" {
		text = fmt.Sprintf("\nERROR: matrix flag host format not valid. %v.\n", err)
		return text
	}
	return text
}

func validateRoom(room string, host string) string {
	var text string
	u := strings.Split(host, "//")[1]
	suffix := fmt.Sprintf(":%s", u)
	if !(strings.HasSuffix(room, suffix)) {
		text = "\nERROR: matrix flag room format is not valid. missing matrix hostname.\n"
		return text
	}
	if !(strings.HasPrefix(room, "#")) && !(strings.HasPrefix(room, "!")) {
		text = "\nERROR: matrix flag room format is not valid. missing room/alias prefix.\n"
		return text
	}
	return text
}

func (r rootVars) rootVarsPresent() string {
	var text string
	v := reflect.ValueOf(r)
	typeOfR := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if len(v.Field(i).Interface().(string)) == 0 {
			text += fmt.Sprintf("\nERROR: matrix flag %s required.\n", strings.ToLower(typeOfR.Field(i).Name))
			text += fmt.Sprintf("set it via the environment variable TRIX_%s or the command line flag --%s.\n", strings.ToUpper(typeOfR.Field(i).Name), strings.ToLower(typeOfR.Field(i).Name))
		}
	}
	if len(r.Host) > 0 {
		text += validateHost(r.Host)
	}
	if len(r.Room) > 0 {
		text += validateRoom(r.Room, r.Host)
	}
	return text
}
