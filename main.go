package main

// TODO
// Maybe filter out notification sounds for like discord
// Add sending a signal to choose your active player

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const intro = `
<node>
	<interface name="com.ornyx.MprisWatcher">
   		<method name="Next">
   		</method>
   		<method name="Prev">
   		</method>
   	</interface>` + introspect.IntrospectDataString + `</node> `

type player struct {
	dbus.BusObject
	name string
}

// Next go to the next active mediaplayer
func (p *player) Next() *dbus.Error {
	if len(PlayerList) == 0 {
		PlayerNum = 0
		return nil
	}
	
	PlayerNum = (PlayerNum+1) % len(PlayerList)
	updateMediaMap()
	return nil
}

// Prev go to the previous active mediaplayer
func (p *player) Prev() *dbus.Error {
	if len(PlayerList) == 0 {
		PlayerNum = 0
		return nil
	}
	
	PlayerNum = (PlayerNum + len(PlayerList) - 1) % len(PlayerList)
	updateMediaMap()
	return nil
}

var Conn *dbus.Conn
var PlayerList []*player
var ActivePlayer *player
var PlayerNum = 0 // the number in PlayerList to use

// the play pause constants - TODO - maybe replace these with symbols from like a config file or smth
const (
	PlaySymbol  = "⏵"
	PauseSymbol = "⏸"
	StopSymbol  = "⏹"
)

func init() {
	var err error
	Conn, err = dbus.ConnectSessionBus()
	failFunc := func(err error) {
		Conn.Close()
		log.Fatalf("Failed to initialize: %v", err)
	}

	if err != nil {
		failFunc(fmt.Errorf("Failed to connect to dbus session bus: %v", err))
	}

	err = updateMediaMap()
	if err != nil {
		failFunc(fmt.Errorf("Failed to update initial media map: %v", err))
	}

	err = initMatchSignals()
	if err != nil {
		failFunc(fmt.Errorf("Failed to initialize the match signals: %v", err))
	}

	err = initMethodExport()
	if err != nil {
		failFunc(err)
	}
}

// initMatchSignals initialize the match signals
// TODO - make it only match mediaplayer signals
func initMatchSignals() error {
	err := Conn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.DBus"))
	if err != nil {
		return fmt.Errorf("Failed to add properties match signal: %v", err)
	}

	err = Conn.AddMatchSignal(
		dbus.WithMatchMember("PropertiesChanged"))
	if err != nil {
		return fmt.Errorf("Failed to add properties match signal: %v", err)
	}

	return nil
}

// initMethodExport initialize the method exports for the Player Object
func initMethodExport() error {
	Conn.Export(ActivePlayer, "/com/ornyx/MprisWatcher", "com.ornyx.MprisWatcher")
	Conn.Export(introspect.Introspectable(intro), "/com/ornyx/MprisWatcher",
		"org.freedesktop.DBus.Introspectable")

	reply, err := Conn.RequestName("com.ornyx.MprisWatcher", dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("Name already taken, do you have multiple of this program open?")
	}

	return nil
}

func main() {
	defer Conn.Close()

	c := make(chan *dbus.Signal, 20)
	Conn.Signal(c)
	c <- nil // send fake signal just so it does the initial print

	for {
		dbusName, identity, playbackStatus, title, shouldchange := stepProg(c)
		if shouldchange {	
			fmt.Printf("%s\x1f%s\x1f%s\x1f%s\n", dbusName, identity, playbackStatus, title)
			// for _, p := range PlayerList {
			// 	fmt.Println(p.GetPlayerInfo())
			// }

			// fmt.Println(PlayerNum)
			// fmt.Println()
		}
	}
}

// stepProg step the program once and return the player information
func stepProg(c chan *dbus.Signal) (dbusName, niceName, playbackStatus, title string, shouldChange bool) {
	select {
	case s := <-c:
		if s != nil && s.Body != nil {
			if s.Path == "/org/freedesktop/Notifications" { // don't update now
				return
			}

			body, ok := s.Body[0].(string)
			if ok && (strings.Contains(body, "MediaPlayer2")) {
				updateMediaMap()
			}
		}

		// fmt.Printf("\n%+v  |  ", s)

		if ActivePlayer == nil {
			shouldChange = true
			return
		}
		
		track, playbackStatus, identity, _ := ActivePlayer.GetPlayerInfo()
		return ActivePlayer.name, identity, playbackStatus, track, true
	}
}

// updateMediaMap Scan the dbus items and update the MediaMap with any
// that don't exist there, also check that none have been removed
func updateMediaMap() error {
	var s []string
	var err error
	var play *player
	var newPlayerList = []*player{}

	err = Conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&s) // get all the objects
	if err != nil {
		Conn.Close()
		log.Fatalln("Failed to get dbus names: ", err)
	}

	for _, v := range s {
		if strings.Contains(v, "MediaPlayer2") {
			play = &player{BusObject: Conn.Object(v, "/org/mpris/MediaPlayer2"), name: v}
			newPlayerList =  append(newPlayerList, play)

			if ActivePlayer == nil {
				ActivePlayer = newPlayerList[0]
			}
		}
	}

	PlayerList = newPlayerList
	if len(PlayerList) == 0 { // set player to nil if no players found
		ActivePlayer = nil
		return nil
	}

	PlayerNum = PlayerNum % len(PlayerList)
	ActivePlayer = PlayerList[PlayerNum]

	return nil
}

// GetPlayerInfo get the player info
func (player *player) GetPlayerInfo() (track, playbackStatus, identity string, err error) {
	track, err = player.getPlayerTrack()
	if err != nil {
		
	}

	playbackStatus, err = player.getPlaybackStatus()
	if err != nil {
		
	}

	identity, err = player.getPlayerIdentity()
	if err != nil {
		
	}

	return
}

// getPlayerTrack Take in a busobject and return its currently playing track and it's playback status
func (player *player) getPlayerTrack() (string, error) {
	prop, err := player.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return "", fmt.Errorf("Failed to get property: %v", err)
	}

	v, ok := prop.Value().(map[string]dbus.Variant)
	if ok {
		if val, ok := v["xesam:title"]; ok {
			return val.String(), nil
		}
		if val, ok := v["xesam:url"]; ok {
			return val.String(), nil
		}
		return "", fmt.Errorf("Media missing track title")
	}

	return "", fmt.Errorf("Metadata not found")
}

// getPlaybackStatus get the current playback status for the player object
// the string value will be based on the constants
func (player *player) getPlaybackStatus() (string, error) {
	prop, err := player.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		return "", fmt.Errorf("Failed to get property: %v", err)
	}

	status, ok := prop.Value().(string)
	if ok {
		switch status {
		case "Playing":
			return PlaySymbol, nil
		case "Paused":
			return PauseSymbol, nil
		case "Stopped":
			return StopSymbol, nil
		}
	}

	return "", fmt.Errorf("Failed to match playback status with value %v", status)
}

// getPlayerIdentity Get the identity for the given player
func (player *player) getPlayerIdentity() (string, error) {
	prop, err := player.GetProperty("org.mpris.MediaPlayer2.Identity")
	if err != nil {
		return "", fmt.Errorf("Failed to get identity: %v", err)
	}

	identity, ok := prop.Value().(string)
	if ok {
		return identity, nil
	}

	return "", fmt.Errorf("Failed to find identity")
}
