package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	trix "codeberg.org/meh/trix/matrix"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var self trix.MaTrix
var debug string
var botMessage string

// TestMain will exec each test, one by one
func TestMain(m *testing.M) {
	setUp()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

// create matrix room
func createRoom(vis string, alias string, direct bool) *mautrix.RespCreateRoom {
	rm := &mautrix.ReqCreateRoom{
		Visibility:    vis,
		RoomAliasName: alias,
		Invite:        []id.UserID{"@bot:trix.meh"},
		IsDirect:      direct,
	}
	resp, err := self.Client.CreateRoom(rm)
	if errors.Is(err, mautrix.MRoomInUse) {
		log.Debug().Msgf("Room %s already created", alias)
	} else if err != nil {
		log.Error().Stack().Err(err).Msgf("Create room %s", alias)
	}
	return resp
}

// setUp function, add a number to numbers slice
func setUp() {
	var roomRes *mautrix.RespCreateRoom
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	debug, ok := os.LookupEnv("DEBUG")
	if ok {
		if debug == "true" {
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// trix admin user login
	self.MaLogin("http://trix.meh:8008", "trix", "trix")

	// create public room & trix user join
	roomRes = createRoom("public", "public", false)
	log.Debug().Msgf("Create room public: %v\n", roomRes)
	self.MaJoinRoom("#public:trix.meh")

	// create ptivate room & trix user join
	roomRes = createRoom("private", "private", false)
	log.Debug().Msgf("Create room private: %v\n", roomRes)
	self.MaJoinRoom("#private:trix.meh")

	// initialize trix user sql cryptostore & olm machine
	self.MaUserEnc("trix", "trix", "http://trix.meh:8008")

	start := time.Now().UnixNano() / 1_000_000
	trixSyncer := self.Client.Syncer.(*mautrix.DefaultSyncer)
	trixSyncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
		self.Olm.ProcessSyncResponse(resp, since)
		return true
	})
	trixSyncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		self.Olm.HandleMemberEvent(evt)
	})
	trixSyncer.OnEventType(event.EventEncrypted, func(source mautrix.EventSource, evt *event.Event) {
		if evt.Timestamp < start {
			// Ignore events from before the program started
			return
		}
		decrypted, err := self.Olm.DecryptMegolmEvent(evt)
		if err != nil {
			log.Error().Stack().Err(err).Msg("Decrypt message")
		} else {
			log.Debug().Msgf("Received encrypted event: %v", decrypted.Content.Raw)
			message, isMessage := decrypted.Content.Parsed.(*event.MessageEventContent)
			if isMessage {
				botMessage = message.Body
			}
		}
	})
	// start polling in the background
	go func() {
		err := self.Client.Sync()
		if err != nil {
			log.Error().Stack().Err(err).Msg("User trix client sync")
		}
	}()
}

// tearDown function
func tearDown() {
	self.MaLogout()
	self.MaDBclose()
}

//  write an encrypted text message
func TestWriteEncText(t *testing.T) {

	text := "the rain in spain falls mainly on the plain"
	var cmd *exec.Cmd
	if debug == "true" {
		cmd = exec.Command("./trix", "out", "-o", "http://trix.meh:8008", "-u", "bot", "-p", "bot", "-r", "#public:trix.meh", "-t", text, "-v")
	} else {
		cmd = exec.Command("./trix", "out", "-o", "http://trix.meh:8008", "-u", "bot", "-p", "bot", "-r", "#public:trix.meh", "-t", text)
	}
	out, err := cmd.CombinedOutput()
	log.Debug().Msgf("trix cli bot user cmd out:\n%s", string(out))
	if err != nil {
		t.Errorf("Error trix cli bot user cmd.Run() failed: %s", err)
	}
	time.Sleep(5 * time.Second) // give the trix syncer a few seconds to read the message
	if botMessage != text {
		t.Errorf("Error trix client read bot message as: %s. Expected: %s", botMessage, text)
	}
}
