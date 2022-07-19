// Package matrix helpful functions for repetitive tasks in the mautrix go module
package matrix

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite db driver
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// MaTrix struct to hold our objects
type MaTrix struct {
	Client  *mautrix.Client
	DBstore *crypto.SQLCryptoStore
	Olm     *crypto.OlmMachine

	db   *sql.DB
	file string
}

// user full account name. i.e. @<user>:<host>
func toAccount(user string, host string) string {
	u, _ := url.Parse(host)
	final := u.Host
	h, p, _ := net.SplitHostPort(u.Host)
	if len(p) > 0 {
		final = h
	}
	return fmt.Sprintf("@%s:%s", user, final)
}

// convert matrix rooim alias to roomID. or convert room string to RoomID type.
func toRoomID(cli *mautrix.Client, room string) id.RoomID {
	if strings.HasPrefix(room, "#") {
		a := id.RoomAlias(room)
		rm, err := cli.ResolveAlias(a)
		if err != nil {
			log.Error().Msgf("Error resolving alias to room %s: %v", room, err)
			os.Exit(1)
		}
		return rm.RoomID
	}
	return id.RoomID(room)
}

// MaLogin client login to matrix & join room
func (t *MaTrix) MaLogin(host string, user string, pass string) {
	var err error
	t.Client, err = mautrix.NewClient(host, "", "")
	if err != nil {
		log.Error().Msgf("Error creating new matrix client for user %s: %v", user, err)
		os.Exit(1)
	}
	_, err = t.Client.Login(&mautrix.ReqLogin{
		Type:                     "m.login.password",
		Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user},
		Password:                 pass,
		InitialDeviceDisplayName: "trix",
		StoreCredentials:         true,
	})
	if err != nil {
		log.Error().Msgf("Error matrix client login user %s: %v", user, err)
		os.Exit(1)
	}
}

// MaJoinRoom to join a room
func (t *MaTrix) MaJoinRoom(room string) {
	rm := toRoomID(t.Client, room)
	_, err := t.Client.JoinRoomByID(rm)
	if err != nil {
		log.Error().Msgf("Error user %s joining room %s: %v", t.Client.UserID, room, err)
		os.Exit(1)
	}
}

// MaLogout matrix client logout
func (t *MaTrix) MaLogout() *mautrix.RespLogout {
	resp, err := t.Client.Logout()
	if err != nil {
		log.Error().Msgf("Error user %s logout: %v", t.Client.UserID, err)
		os.Exit(1)
	}
	return resp
}

// return a random string of fixed length characters
func randString(length int) string {
	var sRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	ran := make([]rune, length)
	for i := range ran {
		ran[i] = letters[sRand.Intn(len(letters))]
	}
	return string(ran)
}

// MaDBopen create matrix SQL cryptostore
func (t *MaTrix) MaDBopen(user string, host string) {
	t.file = fmt.Sprintf("trix.%s", randString(4))
	log.Debug().Msgf("SQL cryptostore db file for user %s: %v", user, t.file)
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), "tmp")); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Join(os.Getenv("HOME"), "tmp"), os.ModeDir)
		if err != nil {
			log.Error().Msgf("Error creating directory ~/tmpi: %v", err)
			os.Exit(1)
		}
	}
	_, err := os.Create(filepath.Join(os.Getenv("HOME"), "tmp", t.file))
	if err != nil {
		log.Error().Msgf("Error creating file ~/tmp/%s for user %s: %v", t.file, user, err)
		os.Exit(1)
	}
	t.db, err = sql.Open("sqlite3", filepath.Join(os.Getenv("HOME"), "tmp", t.file))
	if err != nil {
		log.Error().Msgf("Error opening sql db ~/tmp/%s for user %s: %v", t.file, user, err)
		os.Exit(1)
	}
	acct := toAccount(user, host)
	pickleKey := []byte("trix_is_for_kids")
	t.DBstore = crypto.NewSQLCryptoStore(t.db, "sqlite3", acct, t.Client.DeviceID, pickleKey, &fakeLogger{})
	err = t.DBstore.CreateTables()
	if err != nil {
		log.Error().Msgf("Error creating SQL cryptostore tables ~/tmp/%s for user %s: %v", t.file, user, err)
		os.Exit(1)
	}
}

//MaDBclose delete matrix SQL cryptostore
func (t *MaTrix) MaDBclose() {
	err := t.db.Close()
	if err != nil {
		log.Error().Msgf("Error closing SQL database ~/tmp/%s for user %s: %v", t.file, t.Client.UserID, err)
		os.Exit(1)
	}
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), "tmp", t.file)); err == nil {
		err = os.Remove(filepath.Join(os.Getenv("HOME"), "tmp", t.file))
		if err != nil {
			log.Error().Msgf("Error removing db file ~/tmp/%s: %v", t.file, err)
			os.Exit(1)
		}
	}
}

// crypto.StateStore implementation that says all rooms are encrypted.
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

// crypto.Logger implementation that just prints to stdout.
type fakeLogger struct{}

var _ crypto.Logger = &fakeLogger{}

func (f fakeLogger) Error(message string, args ...interface{}) {
	//fmt.Printf("[ERROR] "+message+"\n", args...)
	warn, err := zerolog.ParseLevel("warn")
	if err != nil {
		log.Error().Msgf("Error cannot parse zerolog warn log level: %v", err)
		os.Exit(1)
	}
	info, err := zerolog.ParseLevel("info")
	if err != nil {
		log.Error().Msgf("Error cannot parse zerolog info log level: %v", err)
		os.Exit(1)
	}
	if strings.HasPrefix(message, "Error while verifying cross-signing keys") {
		if zerolog.GlobalLevel() == warn || zerolog.GlobalLevel() == info {
			return
		}
		log.Error().Msgf(message, args...)
	} else {
		log.Error().Msgf(message, args...)
		os.Exit(1)
	}
}

func (f fakeLogger) Warn(message string, args ...interface{}) {
	//fmt.Printf("[WARN] "+message+"\n", args...)
	log.Warn().Msgf(message, args...)
}

func (f fakeLogger) Debug(message string, args ...interface{}) {
	//fmt.Printf("[DEBUG] "+message+"\n", args...)
	log.Debug().Msgf(message, args...)

}

func (f fakeLogger) Trace(message string, args ...interface{}) {
	//if strings.HasPrefix(message, "Got membership state event") {
	//	return
	//}
	//fmt.Printf("[TRACE] "+message+"\n", args...)
	log.Trace().Msgf(message, args...)
}

// MaOlm create olm machine
func (t *MaTrix) MaOlm() {
	t.Olm = crypto.NewOlmMachine(t.Client, &fakeLogger{}, t.DBstore, &fakeStateStore{})
	t.Olm.AllowUnverifiedDevices = true
	t.Olm.ShareKeysToUnverifiedDevices = false
	// Load data from the crypto store
	err := t.Olm.Load()
	if err != nil {
		log.Error().Msgf("Error loading Olm machine data to matrix for user %s: %v", t.Client.UserID, err)
		os.Exit(1)
	}
	if t.Olm.CrossSigningKeys == nil {
		t.Olm.CrossSigningKeys, err = t.Olm.GenerateCrossSigningKeys()
		if err != nil {
			log.Error().Msgf("Error generating cross signing keys for user %s: %v", t.Client.UserID, err)
			os.Exit(1)
		}
	}
}

// get members of a room
func getUserIDs(cli *mautrix.Client, roomID id.RoomID) []id.UserID {
	members, err := cli.JoinedMembers(roomID)
	if err != nil {
		log.Error().Msgf("Error retreiving room member list for %s: %v", string(roomID), err)
		os.Exit(1)
	}
	userIDs := make([]id.UserID, len(members.Joined))
	i := 0
	for userID := range members.Joined {
		userIDs[i] = userID
		i++
	}
	return userIDs
}

// SendEncrypted text to a room
func (t *MaTrix) SendEncrypted(room string, text string) id.EventID {
	content := event.MessageEventContent{
		MsgType: "m.notice",
		Body:    text,
	}
	rm := toRoomID(t.Client, room)
	encrypted, err := t.Olm.EncryptMegolmEvent(rm, event.EventMessage, content)
	// These three errors mean we have to make a new Megolm session
	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
		err = t.Olm.ShareGroupSession(rm, getUserIDs(t.Client, rm))
		if err != nil {
			log.Error().Msgf("Error sharing group session to %s: %v", room, err)
			os.Exit(1)
		}
		encrypted, err = t.Olm.EncryptMegolmEvent(rm, event.EventMessage, content)
		if err != nil {
			log.Error().Msgf("Error encrypting message to %s: %v", room, err)
			os.Exit(1)
		}
	}
	resp, err := t.Client.SendMessageEvent(rm, event.EventEncrypted, encrypted)
	if err != nil {
		log.Error().Msgf("Error sending encrypted message to %s: %v", room, err)
		os.Exit(1)
	}
	return resp.EventID
}
