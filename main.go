package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

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
		state := "toStart"
		switch *action {
		case "interative":
			for {
				Interative(conn, &state)

			}
		case "file":
			SendFile(conn, "test.txt")
		}
		done <- struct{}{}
	}()
	<-done
}

func Interative(conn net.Conn, state *string) {
	switch *state {
	case "toStart":
		fmt.Printf("Enter username (blank for use your ip): ")
		var username string
		fmt.Scanf("%s", &username)
		RegisterToServer(conn, username)
		*state = "menu"
	case "menu":
		fmt.Printf("Enter command: \n 1. Subscribe to channel \n 2. Publish to channel \n 3. Exit \n")
		var command int
		fmt.Scanf("%d", &command)
		ClearTerminal()
		switch command {
		case 1:
			fmt.Printf("Enter channel number: ")
			var channel int
			fmt.Scanf("%d", &channel)
			SubscribeChannel(conn, channel)

		case 2:
			fmt.Printf("Enter file path: ")
			var filePath string
			fmt.Scanf("%s", &filePath)
			SendFile(conn, filePath)

		case 3:
			fmt.Printf("Exiting\n")
			os.Exit(0)
		default:
			fmt.Printf("Unknown command %d\n", command)
		}
	}
}

func SendFile(conn net.Conn, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	err = CopyContent(conn, file)
	if err != nil {
		fmt.Printf("File: %s not sendend\n", filePath)
	}
	fmt.Printf("File: %s succesful sendend\n", filePath)
}

func CopyContent(dst io.Writer, readersrc io.Reader) error {
	_, err := io.Copy(dst, readersrc)
	if err != nil {
		return err
	}
	return nil
}

func RegisterToServer(conn net.Conn, username string) {
	if username != "" {
		fmt.Fprintf(conn, "register %s\n", username)
		return
	}
	fmt.Fprintf(conn, "register \n")
}

func SubscribeChannel(conn net.Conn, channel int) {
	fmt.Fprintf(conn, "subscribe %d\n", channel)
}

func Publish(conn net.Conn, channel int, message string) {
	fmt.Fprintf(conn, "publish %d %s\n", channel, message)
}
