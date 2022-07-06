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
			t := root.rootVarsPresent()
			if len(t) > 0 {
				cmd.Help()
				fmt.Printf("%s\n", t)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var out trix.MaTrix
			out.MaLogin(root.Host, root.User, root.Pass, root.Room)
			out.MaDBopen(root.User, root.Host)
			out.MaOlm()

			defer func() {
				resp := out.MaLogout()
				fmt.Printf("logout %v\n", resp)
				out.MaDBclose()
			}()

			// Hook up the OlmMachine into the Matrix client so it receives e2ee keys and other such things.
			syncer := out.Client.Syncer.(*mautrix.DefaultSyncer)
			syncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
				out.Olm.ProcessSyncResponse(resp, since)
				return true
			})
			syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
				out.Olm.HandleMemberEvent(evt)
			})

			// Start long polling in the background
			go func() {
				err := out.Client.Sync()
				if err != nil {
					panic(err)
				}
			}()

			resp := out.SendEncrypted(root.Room, text)
			fmt.Printf("%v\n", resp)

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
