package cmd

import (
	"fmt"
	"os"
	"strings"

	trix "codeberg.org/meh/trix/matrix"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var (
	text string // text to send to the matrix room

	out = &cobra.Command{
		Use:   "out",
		Short: "send text to the matrix channel",
		Long:  `send string output to the matrix channel`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// Default level is info, unless debug flag is present
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
			if debug {
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			}
			var errText strings.Builder
			var help bool
			if len(maUser) == 0 {
				errText.WriteString("Error: matrix username is required\n")
				help = true
			}
			if len(maPass) == 0 {
				errText.WriteString("Error: matrix password is required\n")
				help = true
			}
			if len(maHost) == 0 {
				errText.WriteString("Error: matrix hostname is required\n")
				help = true
			} else {
				err := validateHost(maHost)
				if err != nil {
					errText.WriteString(fmt.Sprintf("Error: %s\n", err))
				}
			}
			if len(maRoom) == 0 {
				errText.WriteString("Error: matrix room name is required\n")
				help = true
			} else {
				err := validateRoom(maRoom, maHost)
				if err != nil {
					errText.WriteString(fmt.Sprintf("Error: %s\n", err))
				}
			}
			if len(text) == 0 {
				errText.WriteString("Error: text to send to the matrix room is required\n")
				help = true
			}
			if help {
				err := cmd.Help()
				if err != nil {
					log.Error().Stack().Err(err).Msg("Cannot execute help menu")
					os.Exit(1)
				}
				fmt.Printf("\n%s", errText.String())
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// initialize matrix struct
			var out trix.MaTrix
			// login to matrix host
			out.MaLogin(maHost, maUser, maPass)
			// join the matrix room
			out.MaJoinRoom(maRoom)
			// create sql cryptostore & olm machine
			out.MaUserEnc(maUser, maPass, maHost)

			// defer logout and dbclose til cli exits
			defer func() {
				resp := out.MaLogout()
				log.Debug().Msgf("logout %v\n", resp)
				out.MaDBclose()
			}()

			// initialize matrix syncer & add our olm machine
			syncer := out.Client.Syncer.(*mautrix.DefaultSyncer)
			syncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
				out.Olm.ProcessSyncResponse(resp, since)
				return true
			})
			syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
				out.Olm.HandleMemberEvent(1, evt)
			})

			// start polling in the background
			go func() {
				err := out.Client.Sync()
				if err != nil {
					log.Error().Stack().Err(err).Msgf("User %s client sync", maUser)
				}
			}()

			// send encrypted message
			resp := out.SendEncrypted(maRoom, text)
			log.Debug().Msgf("Sent Message from %s to room %s EventID %s\n", maUser, maRoom, string(resp))

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	rootCmd.AddCommand(out)
}
