// much of this encryption code is lifted with much respect and admiration to its author
// https://mau.dev/-/snippets/6
package cmd

import (
	"fmt"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

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

func sendEncrypted(mach *crypto.OlmMachine, cli *mautrix.Client, roomID id.RoomID, text string) {
	content := event.MessageEventContent{
		MsgType: "m.text",
		Body:    text,
	}
	encrypted, err := mach.EncryptMegolmEvent(roomID, event.EventMessage, content)
	// These three errors mean we have to make a new Megolm session
	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
		err = mach.ShareGroupSession(roomID, getUserIDs(cli, roomID))
		if err != nil {
			panic(err)
		}
		encrypted, err = mach.EncryptMegolmEvent(roomID, event.EventMessage, content)
	}
	if err != nil {
		panic(err)
	}
	resp, err := cli.SendMessageEvent(roomID, event.EventEncrypted, encrypted)
	if err != nil {
		panic(err)
	}
	fmt.Println("Send response:", resp)
}
