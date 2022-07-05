// Package matrix this cli cobra/viper configuration
// much of this encryption code is lifted with much respect and admiration to its author
// https://mau.dev/-/snippets/6
package matrix

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // comment
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// MaTrix struct
type MaTrix struct {
	Client  *mautrix.Client
	DBstore *crypto.SQLCryptoStore
	Olm     *crypto.OlmMachine
}

// user full account name. i.e. @<user>:<host>
func toAccount(user string, host string) string {
	url := strings.Split(host, "//")[1]
	return fmt.Sprintf("@%s:%s", user, url)
}

// convert matrix rooim alias to roomID
func toRoomID(cli *mautrix.Client, room string) (id.RoomID, bool) {
	if strings.HasPrefix(room, "#") {
		a := id.RoomAlias(room)
		rm, err := cli.ResolveAlias(a)
		if err != nil {
			panic(err)
		}
		return rm.RoomID, true
	}
	return id.RoomID(room), false
}

// MaLogin client login to matrix & join room
func MaLogin(host string, user string, pass string, room string) *mautrix.Client {
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
	rm, _ := toRoomID(client, room)
	_, err = client.JoinRoomByID(rm)
	if err != nil {
		panic(err)
	}
	//defer client.Logout()
	return client
}

func randString(length int) string {
	rand.Seed(time.Now().Unix())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	ran := make([]rune, length)
	for i := range ran {
		ran[i] = letters[rand.Intn(len(letters))]
	}
	return string(ran)
}

// MaDB matrix SQL store
func MaDB(cli *mautrix.Client, user string, host string) *crypto.SQLCryptoStore {
	file := fmt.Sprintf("trix.%s", randString(4))
	// Log Debug print filename
	fmt.Println(file)
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), "tmp")); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Join(os.Getenv("HOME"), "tmp"), os.ModeDir)
		if err != nil {
			panic(err)
		}
	}
	_, err := os.Create(filepath.Join(os.Getenv("HOME"), "tmp", file))
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("sqlite3", filepath.Join(os.Getenv("HOME"), "tmp", file))
	if err != nil {
		panic(err)
	}
	acct := toAccount(user, host)
	pickleKey := []byte("trix_is_for_kids")
	cryptoStore := crypto.NewSQLCryptoStore(db, "sqlite3", acct, cli.DeviceID, pickleKey, &fakeLogger{})
	err = cryptoStore.CreateTables()
	if err != nil {
		panic(err)
	}
	//defer os.Remove(filepath.Join(os.Getenv("HOME"), "tmp", file))
	//defer db.Close()
	return cryptoStore
}

// Simple crypto.StateStore implementation that says all rooms are encrypted.
type fakeStateStore struct{}

var _ crypto.StateStore = &fakeStateStore{}

func (fss *fakeStateStore) IsEncrypted(roomID id.RoomID) bool {
	return true
}

func (fss *fakeStateStore) GetEncryptionEvent(roomID id.RoomID) *event.EncryptionEventContent {
	return &event.EncryptionEventContent{
		Algorithm:              id.AlgorithmMegolmV1,
		RotationPeriodMillis:   7 * 24 * 60 * 60 * 1000,
		RotationPeriodMessages: 100,
	}
}

func (fss *fakeStateStore) FindSharedRooms(userID id.UserID) []id.RoomID {
	return []id.RoomID{}
}

// Simple crypto.Logger implementation that just prints to stdout.
type fakeLogger struct{}

var _ crypto.Logger = &fakeLogger{}

func (f fakeLogger) Error(message string, args ...interface{}) {
	fmt.Printf("[ERROR] "+message+"\n", args...)
}

//Log Debug below

func (f fakeLogger) Warn(message string, args ...interface{}) {
	fmt.Printf("[WARN] "+message+"\n", args...)
}

func (f fakeLogger) Debug(message string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+message+"\n", args...)
}

func (f fakeLogger) Trace(message string, args ...interface{}) {
	if strings.HasPrefix(message, "Got membership state event") {
		return
	}
	fmt.Printf("[TRACE] "+message+"\n", args...)
}

// MaOlm olm machine
func MaOlm(cli *mautrix.Client, store *crypto.SQLCryptoStore) *crypto.OlmMachine {
	mach := crypto.NewOlmMachine(cli, &fakeLogger{}, store, &fakeStateStore{})
	mach.AllowUnverifiedDevices = false
	mach.ShareKeysToUnverifiedDevices = false
	// Load data from the crypto store
	err := mach.Load()
	if err != nil {
		panic(err)
	}
	return mach
}

// Easy way to get room members (to find out who to share keys to).
// In real apps, you should cache the member list somewhere and update it based on m.room.member events.
func getUserIDs(cli *mautrix.Client, roomID id.RoomID) []id.UserID {
	members, err := cli.JoinedMembers(roomID)
	if err != nil {
		panic(err)
	}
	userIDs := make([]id.UserID, len(members.Joined))
	i := 0
	for userID := range members.Joined {
		userIDs[i] = userID
		i++
	}
	return userIDs
}

// SendEncrypted func
func SendEncrypted(cli *mautrix.Client, mach *crypto.OlmMachine, room string, text string) id.EventID {
	content := event.MessageEventContent{
		MsgType: "m.text",
		Body:    text,
	}
	rm, _ := toRoomID(cli, room)
	encrypted, err := mach.EncryptMegolmEvent(rm, event.EventMessage, content)
	// These three errors mean we have to make a new Megolm session
	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
		err = mach.ShareGroupSession(rm, getUserIDs(cli, rm))
		if err != nil {
			panic(err)
		}
		encrypted, err = mach.EncryptMegolmEvent(rm, event.EventMessage, content)
		if err != nil {
			panic(err)
		}
	}
	resp, err := cli.SendMessageEvent(rm, event.EventEncrypted, encrypted)
	if err != nil {
		panic(err)
	}
	return resp.EventID
}
