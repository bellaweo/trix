package main

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	trix "codeberg.org/meh/trix/matrix"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var self trix.MaTrix

// TestMain will exec each test, one by one
func TestMain(m *testing.M) {
	setUp()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

// execute docker exec command to create new user on matrix host
func addUser(cli *client.Client, container string, cmd []string) string {
	trix := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Privileged:   true,
		Tty:          true,
		Cmd:          cmd,
	}

	rst, err := cli.ContainerExecCreate(context.Background(), container, trix)
	if err != nil {
		log.Error().Msgf("Error creating docker exec: %v", err)
		os.Exit(1)
	}

	response, err := cli.ContainerExecAttach(context.Background(), rst.ID, types.ExecStartCheck{})
	if err != nil {
		log.Error().Msgf("Error execuing docker exec: %v", err)
		os.Exit(1)
	}
	defer response.Close()

	data, err := ioutil.ReadAll(response.Reader)
	if err != nil {
		log.Error().Msgf("Error rading docker exec command: %v", err)
		os.Exit(1)
	}
	return string(data)
}

func createRoom(vis string, alias string, direct bool) *mautrix.RespCreateRoom {
	rm := &mautrix.ReqCreateRoom{
		Visibility:    vis,
		RoomAliasName: alias,
		Invite:        []id.UserID{"@bot:localhost"},
		IsDirect:      direct,
	}
	resp, err := self.Client.CreateRoom(rm)
	if errors.Is(err, mautrix.MRoomInUse) {
		log.Debug().Msgf("Room %s already created", alias)
	} else if err != nil {
		log.Error().Msgf("Error creating room %s: %v", alias, err)
		os.Exit(1)
	}
	return resp
}

// setUp function, add a number to numbers slice
func setUp() {
	var ct string
	var cmd []string
	var cmdRes string
	var roomRes *mautrix.RespCreateRoom
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// get the running matrix host container id
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Error().Msgf("Error creating docker client: %v", err)
		os.Exit(1)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error().Msgf("Error getting list of running containers: %v", err)
		os.Exit(1)
	}
	for _, container := range containers {
		if container.Image == "synapse_matrix" {
			ct = container.ID[:10]
		}
	}

	// add the matrix users
	cmd = []string{"register_new_matrix_user", "http://localhost:8008", "-c", "/data/homeserver.yaml", "-u", "trix", "-p", "trix", "-a"}
	cmdRes = addUser(cli, ct, cmd)
	log.Debug().Msgf("Create trix user: %s", cmdRes)

	cmd = []string{"register_new_matrix_user", "http://localhost:8008", "-c", "/data/homeserver.yaml", "-u", "bot", "-p", "bot", "--no-admin"}
	cmdRes = addUser(cli, ct, cmd)
	log.Debug().Msgf("Create bot user: %s", cmdRes)

	// trix admin user login
	self.MaLogin("http://localhost:8008", "trix", "trix")

	// create public room & trix user join
	roomRes = createRoom("public", "public", false)
	log.Debug().Msgf("Create room public: %v\n", roomRes)
	self.MaJoinRoom("#public:localhost")

	// create ptivate room & trix user join
	roomRes = createRoom("private", "private", false)
	log.Debug().Msgf("Create room private: %v\n", roomRes)
	self.MaJoinRoom("#private:localhost")

	self.MaDBopen("trix", "http://localhost:8008")
	self.MaOlm()

	trixSyncer := self.Client.Syncer.(*mautrix.DefaultSyncer)
	trixSyncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
		self.Olm.ProcessSyncResponse(resp, since)
		return true
	})
	trixSyncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		self.Olm.HandleMemberEvent(evt)
	})
	// start polling in the background
	go func() {
		err := self.Client.Sync()
		if err != nil {
			log.Error().Msgf("Error trix user matrix sync: %v", err)
			os.Exit(1)
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
	var bot trix.MaTrix
	bot.MaLogin("http://localhost:8008", "bot", "bot")
	defer bot.MaLogout()
	bot.MaJoinRoom("#public:localhost")
	bot.MaDBopen("bot", "http://localhost:8008")
	defer bot.MaDBclose()
	bot.MaOlm()
	botSyncer := bot.Client.Syncer.(*mautrix.DefaultSyncer)
	botSyncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
		bot.Olm.ProcessSyncResponse(resp, since)
		return true
	})
	botSyncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		bot.Olm.HandleMemberEvent(evt)
	})

	// start polling in the background
	go func() {
		err := bot.Client.Sync()
		if err != nil {
			log.Error().Msgf("Error bot user matrix sync: %v", err)
			os.Exit(1)
		}
	}()

	// send encrypted message
	resp := bot.SendEncrypted("#public:localhost", "test message from the robot")
	log.Debug().Msgf("Sent Message from bot to room #public:localhost EventID %v\n", resp)
	//	numbers["one"] = numbers["one"] + 1
	//
	//	if numbers["one"] != 2 {
	//		t.Error("1 plus 1 = 2, not %v", value)
	//	}
}
