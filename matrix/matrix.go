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
	"maunium.net/go/mautrix/crypto/ssss"
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
			log.Error().Stack().Err(err).Msg("Resolve alias to room")
		}
		return rm.RoomID
	}
	return id.RoomID(room)
}

// MaLogin client login to matrix
func (t *MaTrix) MaLogin(host string, user string, pass string) {
	var err error
	t.Client, err = mautrix.NewClient(host, "", "")
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Create new matrix client for user %s", user)
	}
	_, err = t.Client.Login(&mautrix.ReqLogin{
		Type:                     "m.login.password",
		Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: user},
		Password:                 pass,
		InitialDeviceDisplayName: "trix",
		StoreCredentials:         true,
	})
	if err != nil {
		log.Error().Stack().Err(err).Msgf("User %s log into matrix", user)
	}
}

// MaJoinRoom to join a room
func (t *MaTrix) MaJoinRoom(room string) {
	rm := toRoomID(t.Client, room)
	_, err := t.Client.JoinRoomByID(rm)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("User %s join room %s", string(t.Client.UserID), room)
	}
}

// MaLogout matrix client logout
func (t *MaTrix) MaLogout() *mautrix.RespLogout {
	resp, err := t.Client.Logout()
	if err != nil {
		log.Error().Stack().Err(err).Msgf("User %s logout of matrix", string(t.Client.UserID))
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

// crypto.Logger implementation that logs to zerolog
type fakeLogger struct{}

var _ crypto.Logger = &fakeLogger{}

func (f fakeLogger) Error(message string, args ...interface{}) {
	warn, err := zerolog.ParseLevel("warn")
	if err != nil {
		log.Error().Stack().Err(err).Msg("Parse zerolog warn level")
	}
	info, err := zerolog.ParseLevel("info")
	if err != nil {
		log.Error().Stack().Err(err).Msg("Parse zerolog info level")
	}
	if strings.HasPrefix(message, "Error while verifying cross-signing keys") {
		if zerolog.GlobalLevel() == warn || zerolog.GlobalLevel() == info {
			return
		}
		log.Error().Stack().Msgf(message, args...)
	} else {
		log.Error().Stack().Msgf(message, args...)
	}
}

func (f fakeLogger) Warn(message string, args ...interface{}) {
	log.Warn().Msgf(message, args...)
}

func (f fakeLogger) Debug(message string, args ...interface{}) {
	log.Debug().Msgf(message, args...)

}

func (f fakeLogger) Trace(message string, args ...interface{}) {
	log.Trace().Msgf(message, args...)
}

// MaUserEnc create matrix SQL cryptostorea & user olm machine
func (t *MaTrix) MaUserEnc(user string, pass string, host string) {
	// create the sql cryptostore (sqlite)
	t.file = fmt.Sprintf("trix.%s", randString(4))
	log.Debug().Msgf("SQL cryptostore db file for user %s: %v", user, t.file)
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), "tmp")); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Join(os.Getenv("HOME"), "tmp"), os.ModeDir)
		if err != nil {
			log.Error().Stack().Err(err).Msg("Create directory ~/tmp")
		}
	}
	_, err := os.Create(filepath.Join(os.Getenv("HOME"), "tmp", t.file))
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Create file ~/tmp/%s for user %s", t.file, string(t.Client.UserID))
	}
	t.db, err = sql.Open("sqlite3", filepath.Join(os.Getenv("HOME"), "tmp", t.file))
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Open sql crypto db ~/tmp/%s for user %s", t.file, string(t.Client.UserID))
	}
	acct := toAccount(user, host)
	pickleKey := []byte("trix_is_for_kids")
	t.DBstore = crypto.NewSQLCryptoStore(t.db, "sqlite3", acct, t.Client.DeviceID, pickleKey, &fakeLogger{})
	err = t.DBstore.CreateTables()
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Create sql crypto db tables ~/tmp/%s for user %s", t.file, string(t.Client.UserID))
	}
	// create the olm machine
	t.Olm = crypto.NewOlmMachine(t.Client, &fakeLogger{}, t.DBstore, &fakeStateStore{})
	t.Olm.AllowUnverifiedDevices = true
	t.Olm.ShareKeysToUnverifiedDevices = false
	// check if SSSS keys already exist for this user. if not, generate & upload
	key, err := t.Olm.SSSS.GetDefaultKeyID()
	if err != nil && err != ssss.ErrNoDefaultKeyAccountDataEvent {
		log.Error().Stack().Err(err).Msgf("Retrieve default SSSS key for user %s", string(t.Client.UserID))
	}
	if len(key) == 0 {
		_, err = t.Olm.GenerateAndUploadCrossSigningKeys(pass, "trix is for kids")
		if err != nil {
			log.Error().Stack().Err(err).Msgf("Create and upload SSSS cross signing keys for user %s", string(t.Client.UserID))
		}
	}
	// Load data from the crypto store
	err = t.Olm.Load()
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Load Olm machine to matrix for user %s", string(t.Client.UserID))
	}
}

//MaDBclose delete matrix SQL cryptostore
func (t *MaTrix) MaDBclose() {
	err := t.db.Close()
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Close sql crypto db ~/tmp/%s for user %s", t.file, string(t.Client.UserID))
	}
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), "tmp", t.file)); err == nil {
		err = os.Remove(filepath.Join(os.Getenv("HOME"), "tmp", t.file))
		if err != nil {
			log.Error().Stack().Err(err).Msgf("Remove sql crypto store db file ~/tmp/%s", t.file)
		}
	}
}

// get members of a room
func getUserIDs(cli *mautrix.Client, roomID id.RoomID) []id.UserID {
	members, err := cli.JoinedMembers(roomID)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Retrieve room memebr list for %s", string(roomID))
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
			log.Error().Stack().Err(err).Msgf("Share group session to room %s", string(rm))
		}
		encrypted, err = t.Olm.EncryptMegolmEvent(rm, event.EventMessage, content)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("Encrypt message to room %s", rm)
		}
	}
	resp, err := t.Client.SendMessageEvent(rm, event.EventEncrypted, encrypted)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Send encrypted message to room %s", rm)
	}
	return resp.EventID
}
