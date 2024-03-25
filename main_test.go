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

	c := make(chan *dbus.Signal, 20)
	Conn.Signal(c)
	c <- nil
	
	for time.Since(testStartTime) < totalTestTime {
		d, n, p, t, change := stepProg(c)
		if change {
			log.Printf("%s\x1f%s\x1f%s\x1f%s\n", d, n, p, t)
		}
	}
}

func BenchmarkProg(b *testing.B) {
	testStartTime := time.Now()
	defer Conn.Close()

	c := make(chan *dbus.Signal, 20)
	Conn.Signal(c)
	c <- nil
	
	for time.Since(testStartTime) < totalTestTime {
		stepProg(c)
	}
}
