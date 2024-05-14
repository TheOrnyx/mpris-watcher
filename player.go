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

// PlayPause toggle between playing and pausing
func (p *player) PlayPause() *dbus.Error {
	if p == nil {
		return nil
	}

	p.Call("org.mpris.MediaPlayer2.Player.PlayPause", 0)
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

	identity, err = player.getPlayerDesktopEntry()
	if err != nil {

	}
	
	// identity, err = player.getPlayerIdentity()
	// if err != nil {

	// }

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
