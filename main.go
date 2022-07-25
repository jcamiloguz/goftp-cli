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
	FILE
	SUB
	UNSUB
	OK
	ERR
)

type Action struct {
	Id      ACTIONID
	Args    map[string]string
	payload []byte
}

var (
	port     = flag.Int("p", 3090, "port")
	host     = flag.String("h", "localhost", "host")
	action   = flag.String("a", "interative", "action")
	channel  = flag.Int("c", 0, "channel")
	filePath = flag.String("f", "", "file path")
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
			// register to server and publish to channel, then exit

			err := RegisterToServer(conn)
			if err != nil {
				log.Fatal(err)
			}
			err = SendFile(conn, *filePath, *channel)
			if err != nil {
				log.Fatal(err)
			}
			done <- struct{}{}

		case "subscribe":

			err := RegisterToServer(conn)
			if err != nil {
				log.Fatal(err)
			}
			err = SubscribeToChannel(conn, *channel)
			if err != nil {
				log.Fatal(err)
			}
			for {

				err = HearingChannel(conn)
				if err != nil {
					errMsg := fmt.Sprintf("Error: %s\n", err)
					log.Fatal(errMsg)
				}
			}
		default:
			log.Fatal("unknown action")
		}
		done <- struct{}{}
	}()
	<-done
}
