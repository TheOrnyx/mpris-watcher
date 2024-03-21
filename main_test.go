package main

import (
	"log"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

const totalTestTime = time.Second * 20 // the total test time in seconds

func Test(t *testing.T) {
	testStartTime := time.Now()
	defer Conn.Close()

	err := Conn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.DBus"))
	if err != nil {
		log.Fatalf("Failed to add properties match signal: %v", err)
	}
	c := make(chan *dbus.Signal, 20)
	Conn.Signal(c)
	c <- nil
	
	for time.Since(testStartTime) < totalTestTime {
		stepProg(c)
	}
}

func BenchmarkProg(b *testing.B) {
	testStartTime := time.Now()
	defer Conn.Close()

	err := Conn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.DBus"))
	if err != nil {
		log.Fatalf("Failed to add properties match signal: %v", err)
	}
	c := make(chan *dbus.Signal, 20)
	Conn.Signal(c)
	c <- nil
	
	for time.Since(testStartTime) < totalTestTime {
		stepProg(c)
	}
}
