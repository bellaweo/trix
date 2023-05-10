// Package cmd trix implementation of cobra
package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	maUser  string
	maPass  string
	maHost  string
	maRoom  string
	debug   bool
	version string

	rootCmd = &cobra.Command{
		Use:     "trix",
		Version: version,
		Short:   "matrix cli",
		Long:    "a matrix cli for performing one-off tasks.",
		PreRun: func(cmd *cobra.Command, args []string) {
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
			err := cmd.Help()
			if err != nil {
				log.Error().Stack().Err(err).Msg("Cannot execute help menu")
				os.Exit(1)
			}
			os.Exit(0)
		},
	}
)

// Execute run the damn thing
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&maUser, "user", "u", "", "matrix username")
	rootCmd.PersistentFlags().StringVarP(&maPass, "pass", "p", "", "matrix password")
	rootCmd.PersistentFlags().StringVarP(&maHost, "host", "o", "", "matrix hostname")
	rootCmd.PersistentFlags().StringVarP(&maRoom, "room", "r", "", "matrix roomid or alias")
	rootCmd.PersistentFlags().BoolVarP(&debug, "verbose", "v", false, "enable verbose mode")
}

// validate the maHost input
func validateHost(host string) error {
	_, err := url.ParseRequestURI(host)
	if err != nil {
		return fmt.Errorf("matrix hostname input format not valid: %s", err)
	}
	u, err := url.Parse(host)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("matrix hostname input format not valid: %s", err)
	}
	return nil
}

// validate the maRoom input
func validateRoom(room string, host string) error {
	u, _ := url.Parse(host)
	suffix := fmt.Sprintf(":%s", u.Host)
	h, p, _ := net.SplitHostPort(u.Host)
	if len(p) > 0 {
		suffix = fmt.Sprintf(":%s", h)
	}
	if !(strings.HasSuffix(room, suffix)) {
		return fmt.Errorf("matrix room input format not valid: missing matrix hostname")
	}
	if !(strings.HasPrefix(room, "#")) && !(strings.HasPrefix(room, "!")) {
		return fmt.Errorf("matrix room input format not valid: missing room/alias prefix")
	}
	return nil
}
