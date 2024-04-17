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

// Next go to the next active mediaplayer
func (p *player) Next() *dbus.Error {
	if len(PlayerList) == 0 {
		PlayerNum = 0
		return nil
	}

	PlayerNum = (PlayerNum + 1) % len(PlayerList)
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
