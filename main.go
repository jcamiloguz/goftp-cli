package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

type ACTIONID int

const (
	REG ACTIONID = iota
	OUT
	PUB
	SUB
	UNSUB
	INFO
	OK
	ERR
)

type Action struct {
	Id   ACTIONID
	Args map[string]string
}

var (
	port   = flag.Int("p", 3090, "port")
	host   = flag.String("h", "localhost", "host")
	action = flag.String("a", "interative", "action")
)

func main() {
	flag.Parse()
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	done := make(chan struct{})
	go func() {
		switch *action {
		case "interative":
			Interative(conn)
		case "publish":
			// Publish(conn, "test", "test")
		case "subscribe":
			// Subscribe(conn, "test")
		default:

		}
		done <- struct{}{}
	}()
	<-done
}
