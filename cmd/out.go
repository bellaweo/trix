package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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

			client, err := mautrix.NewClient(host, "", "")
			if err != nil {
				panic(err)
			}
			_, err = client.Login(&mautrix.ReqLogin{
				Type:                     "m.login.password",
				Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user},
				Password:                 pass,
				InitialDeviceDisplayName: "trix",
				StoreCredentials:         true,
			})
			if err != nil {
				panic(err)
			}

			// Create a store for the e2ee keys. In real apps, use NewSQLCryptoStore instead of NewGobStore.
			cryptoStore, err := crypto.NewGobStore("test.gob")
			if err != nil {
				panic(err)
			}

			mach := crypto.NewOlmMachine(client, &fakeLogger{}, cryptoStore, &fakeStateStore{})
			// Load data from the crypto store
			err = mach.Load()
			if err != nil {
				panic(err)
			}

			// Hook up the OlmMachine into the Matrix client so it receives e2ee keys and other such things.
			syncer := client.Syncer.(*mautrix.DefaultSyncer)
			syncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
				mach.ProcessSyncResponse(resp, since)
				return true
			})
			syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
				mach.HandleMemberEvent(evt)
			})

			// Start long polling in the background
			go func() {
				err = client.Sync()
				if err != nil {
					panic(err)
				}
			}()

			rm := id.RoomID(room)

			go sendEncrypted(mach, client, rm, text)

			_, err = client.Logout()
			if err != nil {
				panic(err)
			}

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
