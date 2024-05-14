package main

// TODO
// Maybe filter out notification sounds for like discord
// Add sending a signal to choose your active player
// definitely fix the signal filtering

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

var intro string

var Conn *dbus.Conn
var PlayerList []*player
var ActivePlayer *player
var introspectObj = &introspector{}
var PlayerNum = 0                 // the number in PlayerList to use
var DisplayInfo = new(activeInfo) // the display info for the current player (updated constantly)

// the play pause constants - TODO - maybe replace these with symbols from like a config file or smth
const (
	PlaySymbol  = "⏵"
	PauseSymbol = "⏸"
	StopSymbol  = "⏹"
)

// initialize a bunch of stuff for the script
func init() {
	var err error
	Conn, err = dbus.ConnectSessionBus()
	failFunc := func(err error) {
		Conn.Close()
		log.Fatalf("Failed to initialize: %v", err)
	}

	introFile, err := os.ReadFile("./watcher-introspect.xml")
	if err != nil {
		failFunc(fmt.Errorf("Failed to open introspect xml: %v", err))
	}
	intro = string(introFile)

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
	Conn.Export(introspectObj, "/com/ornyx/MprisWatcher", "com.ornyx.MprisWatcher")
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
		shouldchange := stepProg(c, DisplayInfo)
		if shouldchange {
			fmt.Printf("%s\n", DisplayInfo)
			// for _, p := range PlayerList {
			// 	fmt.Println(p.GetPlayerInfo())
			// }

			// fmt.Println(PlayerNum)
			// fmt.Println()
		}
	}
}

// stepProg step the program once and return the player information
func stepProg(c chan *dbus.Signal, info *activeInfo) (shouldChange bool) {
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
			DisplayInfo.Clear()
			return
		}

		ActivePlayer.GetPlayerInfo(info) // put all the activeplayers info into the info struct
		return true
	}
}

// updateMediaMap Scan the dbus items and update the MediaMap with any that don't exist
// there, also check that none have been removed
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
			newPlayerList = append(newPlayerList, play)

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
