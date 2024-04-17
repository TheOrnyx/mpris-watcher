package main

import "github.com/godbus/dbus/v5"

type introspector struct{}

// Next go to the next active mediaplayer
func (i *introspector) Next() *dbus.Error {
	if len(PlayerList) == 0 {
		PlayerNum = 0
		return nil
	}

	PlayerNum = (PlayerNum + 1) % len(PlayerList)
	updateMediaMap()
	return nil
}

// Prev go to the previous active mediaplayer
func (i *introspector) Prev() *dbus.Error {
	if len(PlayerList) == 0 {
		PlayerNum = 0
		return nil
	}

	PlayerNum = (PlayerNum + len(PlayerList) - 1) % len(PlayerList)
	updateMediaMap()
	return nil
}

// PlayPause toggle between playing and pausing on the active player
func (i *introspector) PlayPause() *dbus.Error {
	ActivePlayer.PlayPause()
	return nil
}
