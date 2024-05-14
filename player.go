package main

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// item that represents a dbus BusObject for a mediaplayer
type player struct {
	dbus.BusObject
	name string
}

// activeInfo The struct for containing information about the current
// player Much easier than invoking the methods at runtime and MUCH
// better than passing around 90 strings
type activeInfo struct {
	fullName       string // the FullName of the activeplayer - basically just the structs name field
	identity       string // the identity of the active player
	desktopEntry   string // the desktopEntry for the active player - basically a nicer formatted identity
	playbackStatus string // The PlaybackStatus of the active player, formatted using the symbols
	trackTitle     string // The title of the current playing track, if no metadata is found it's just the link
}

// String String representation of the info
func (a *activeInfo) String() string {
	return fmt.Sprintf("%s\x1f%s\x1f%s\x1f%s", a.fullName, a.desktopEntry, a.playbackStatus, a.trackTitle)
}

// Clear - Clear the activeInfo's fields (used for when activeplayer is nil for example)
// Kinda gross but don't want to make new instances because trying to keep this low memory
func (a *activeInfo) Clear() {
	a.fullName, a.identity = "", ""
	a.desktopEntry, a.playbackStatus = "", ""
	a.trackTitle = ""
}

// PlayPause toggle between playing and pausing
func (p *player) PlayPause() *dbus.Error {
	if p == nil {
		return nil
	}

	p.Call("org.mpris.MediaPlayer2.Player.PlayPause", 0)
	return nil
}

// GetPlayerInfo get the player info and assign it to a given activeInfo object
// Don't really need to check the errors as like they only indicate
// something wasn't found which isn't a big problem
func (player *player) GetPlayerInfo(info *activeInfo) (err error) {
	info.fullName = player.name
	info.trackTitle, err = player.getPlayerTrack()
	info.playbackStatus, err = player.getPlaybackStatus()
	info.desktopEntry, err = player.getPlayerDesktopEntry()
	info.identity, err = player.getPlayerIdentity()
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
			return val.Value().(string), nil // should be fine with the unsafe cast
		}
		if val, ok := v["xesam:url"]; ok {
			return val.Value().(string), nil // it's probably always a string - cbf checking :P
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

// getPlayerDesktopEntry Get this players desktop entry
func (player *player) getPlayerDesktopEntry() (string, error) {
	prop, err := player.GetProperty("org.mpris.MediaPlayer2.DesktopEntry")
	if err != nil {
		return "", fmt.Errorf("Failed to get desktopEntry: %v", err)
	}

	identity, ok := prop.Value().(string)
	if ok {
		return identity, nil
	}

	return "", fmt.Errorf("Failed to find DesktopEntry")
}
