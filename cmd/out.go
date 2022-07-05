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
			out.Client = trix.MaLogin(host, user, pass, room)
			out.DBstore = trix.MaDB(out.Client, user, host)
			out.Olm = trix.MaOlm(out.Client, out.DBstore)

			//client, err := mautrix.NewClient(host, "", "")
			//if err != nil {
			//	panic(err)
			//}
			//_, err = client.Login(&mautrix.ReqLogin{
			//	Type:                     "m.login.password",
			//	Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user},
			//	Password:                 pass,
			//	InitialDeviceDisplayName: "trix",
			//	StoreCredentials:         true,
			//})
			//if err != nil {
			//	panic(err)
			//}
			//rm := id.RoomID(room)
			//_, err = client.JoinRoomByID(rm)
			//if err != nil {
			//	panic(err)
			//}

			// Create a store for the e2ee keys.
			//_, err = os.Create(filepath.Join(os.Getenv("HOME"), "tmp", "out.db"))
			//if err != nil {
			//	panic(err)
			//}
			//db, err := sql.Open("sqlite3", filepath.Join(os.Getenv("HOME"), "tmp", "out.db"))
			//if err != nil {
			//	panic(err)
			//}

			//acct := userToAccount(user, host)
			//pickleKey := []byte("trix_is_for_kids")
			//cryptoStore := crypto.NewSQLCryptoStore(db, "sqlite3", acct, client.DeviceID, pickleKey, &fakeLogger{})

			//err = cryptoStore.CreateTables()
			//if err != nil {
			//	panic(err)
			//}

			//mach := crypto.NewOlmMachine(out.Client, &fakeLogger{}, out.DBStore, &fakeStateStore{})
			//mach.AllowUnverifiedDevices = false
			//mach.ShareKeysToUnverifiedDevices = false
			// Load data from the crypto store
			//err = mach.Load()
			//if err != nil {
			//	panic(err)
			//}

			// load room devices into cryptoStore
			//for _, self := range getUserIDs(client, rm) {
			//	mach.LoadDevices(self)
			//}

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

			resp := trix.SendEncrypted(out.Client, out.Olm, room, text)
			fmt.Printf("%v\n", resp)

			//content := event.MessageEventContent{
			//	MsgType: "m.text",
			//	Body:    text,
			//}
			//encrypted, err := mach.EncryptMegolmEvent(rm, event.EventMessage, content)
			// These three errors mean we have to make a new Megolm session
			//if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
			//	err = mach.ShareGroupSession(rm, getUserIDs(client, rm))
			//	if err != nil {
			//		panic(err)
			//	}
			//	encrypted, err = mach.EncryptMegolmEvent(rm, event.EventMessage, content)
			//	if err != nil {
			//		panic(err)
			//	}
			//}
			//_, err = client.SendMessageEvent(rm, event.EventEncrypted, encrypted)
			//if err != nil {
			//	panic(err)
			//}

			//go sendEncrypted(mach, client, cryptoStore, rm, text)

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
