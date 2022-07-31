package cmd

import (
	"fmt"
	"os"

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
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if debug {
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			}
			// validate flags & values
			t := root.rootVarsPresent()
			if len(t) > 0 {
				cmd.Help()
				fmt.Printf("%s\n", t)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// initialize matrix struct
			var out trix.MaTrix
			// login to matrix host
			out.MaLogin(root.Host, root.User, root.Pass)
			// join the matrix room
			out.MaJoinRoom(root.Room)
			// create sql cryptostore & olm machine
			out.MaUserEnc(root.User, root.Pass, root.Host)

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
				out.Olm.HandleMemberEvent(evt)
			})

			// start polling in the background
			go func() {
				err := out.Client.Sync()
				if err != nil {
					log.Error().Stack().Err(err).Msgf("User %s client sync", root.User)
				}
			}()

			// send encrypted message
			resp := out.SendEncrypted(root.Room, text)
			log.Debug().Msgf("Sent Message from %s to room %s EventID %s\n", root.User, root.Room, string(resp))

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
