package cmd

import (
	"fmt"
	"os"

	trix "codeberg.org/meh/trix/matrix"
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
			out.MaLogin(root.Host, root.User, root.Pass, root.Room)
			// create sql cryptostore
			out.MaDBopen(root.User, root.Host)
			// create olm machine
			out.MaOlm()

			// defer logout and dbclose til cli exits
			defer func() {
				resp := out.MaLogout()
				fmt.Printf("logout %v\n", resp)
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
					panic(err)
				}
			}()

			// send encrypted message
			resp := out.SendEncrypted(root.Room, text)
			fmt.Printf("Sent Message EventID %v\n", resp)

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
