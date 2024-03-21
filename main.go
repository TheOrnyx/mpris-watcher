package main

// TODO
// Maybe filter out notification sounds for like discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

type player struct {
	dbus.BusObject
	name string
}

var Conn *dbus.Conn
var PlayerMap = make(map[string]*player)
var ActivePlayer *player
const ClearLine = "\033[2K\r"

func init() {
	var err error
	Conn, err = dbus.ConnectSessionBus()
	if err != nil {
		log.Fatal("Failed to connect to dbus session bus: ", err)
	}

	err = updateMediaMap()
	if err != nil {
		log.Fatalf("Failed to update initial media map: %v", err)
	}
}

func main() {
	defer Conn.Close()

	err := Conn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.DBus"))
	if err != nil {
		log.Fatalf("Failed to add properties match signal: %v", err)
	}
	c := make(chan *dbus.Signal, 20)
	Conn.Signal(c)
	c <- nil
	
	for {
		stepProg(c)
	}
}

// stepProg step the program once (needed for testing)
func stepProg(c chan *dbus.Signal)  {
	select {
	case s := <-c:
		if s != nil && s.Body != nil {
			body, ok := s.Body[0].(string)
			if ok && (strings.Contains(body, "MediaPlayer2")) {
				updateMediaMap()
			}
		}
		
		// fmt.Printf("\n%+v  |  ", s)
		
		if ActivePlayer == nil {
			fmt.Printf("%sNo media devices",ClearLine)
			return
		}
		
		metadata, err := getPlayerTrack(ActivePlayer)
		if err != nil {
			fmt.Printf("%sNo media playing",ClearLine)
			return
		} else {
			fmt.Printf("%s%s", ClearLine, metadata)
			// fmt.Println(metadata)
		}
	}
}

// updateMediaMap Scan the dbus items and update the MediaMap with any
// that don't exist there, also check that none have been removed
func updateMediaMap() error {
	var s []string
	var err error
	var play *player
	var newPlayerMap = make(map[string]*player)
	
	err = Conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&s) // get all the objects
	if err != nil {
		Conn.Close()
		log.Fatalln("Failed to get dbus names: ", err)
	}

	for _, v := range s {
		if strings.Contains(v, "MediaPlayer2") {

			play = &player{BusObject: Conn.Object(v, "/org/mpris/MediaPlayer2"), name: v}
			newPlayerMap[v] = play

			if ActivePlayer == nil {
				ActivePlayer = newPlayerMap[v]
			}
		}
	}

	PlayerMap = newPlayerMap
	if ActivePlayer == nil {
		return nil
	}
	
	if mapPlay, exists := PlayerMap[ActivePlayer.name]; exists {
		ActivePlayer = mapPlay
	} else {
		ActivePlayer = play
	}
	
	return nil
}

// getPlayerTrack Take in a busobject and return its currently playing track
func getPlayerTrack(player dbus.BusObject) (string, error) {
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
