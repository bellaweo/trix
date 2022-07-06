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
		Run: func(cmd *cobra.Command, args []string) {
			ut, us := vars()
			if ut {
				cmd.Help()
				fmt.Print(us)
				os.Exit(1)
			}

			var out trix.MaTrix
			out.MaLogin(host, user, pass, room)
			out.MaDBopen(user, host)
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

			resp := out.SendEncrypted(room, text)
			fmt.Printf("%v\n", resp)

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
