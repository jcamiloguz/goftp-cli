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
			// SendFile(conn, "test.txt")
		}
		done <- struct{}{}
	}()
	<-done
}

func Interative(conn net.Conn, state *string) {
	var channel int
	switch *state {
	case "toStart":
		fmt.Printf("Enter username (blank for use your ip): ")
		var username string
		fmt.Scanf("%s", &username)
		err := RegisterToServer(conn, username)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		*state = "menu"
	case "menu":
		fmt.Printf("Enter command: \n 1. Subscribe to channel \n 2. Publish to channel \n 3. Exit \n")
		var command int
		fmt.Scanf("%d", &command)
		ClearTerminal()
		switch command {
		case 1:
			fmt.Printf("Enter channel number: ")
			fmt.Scanf("%d", &channel)
			err := SubscribeChannel(conn, channel)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			fmt.Printf("hearing %d channel\n", channel)
			err = HearingChannel(conn)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			os.Exit(0)
			return

		case 2:
			fmt.Printf("Enter file path: ")
			var filePath string
			fmt.Scanf("%s", &filePath)
			var channelToPublish int
			fmt.Printf("Enter channel to publish: ")
			fmt.Scanf("%d", &channelToPublish)
			SendFile(conn, filePath, channelToPublish)
			os.Exit(0)
		case 3:
			fmt.Printf("Exiting\n")
			os.Exit(0)

		default:
			fmt.Printf("Unknown command %d\n", command)
		}

	}
}

func SendFile(conn net.Conn, filePath string, channel int) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("OS. Stat() function execution error, error is:% v \n", err)
		return err
	}
	sizeInKb := fileInfo.Size() / 1024
	fmt.Printf("file %s %dKb\n", fileInfo.Name(), sizeInKb)

	publishMsg := fmt.Sprintf("publish channel=%d  fileName=%s size=%d\n", channel, fileInfo.Name(), sizeInKb)
	_, err = conn.Write([]byte(publishMsg))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	_, err = WaitForResponse(conn)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	err = CopyContent(conn, file)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	return nil
}

func CopyContent(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func RegisterToServer(conn net.Conn, username string) error {
	if username != "" {
		fmt.Fprintf(conn, "register %s\n", username)
	} else {
		fmt.Fprintf(conn, "register \n")
	}
	action, err := WaitForResponse(conn)
	if err != nil {
		return err
	}
	if action.Id != OK {
		return fmt.Errorf("Error expected OK, got %d, with args %v", action.Id, action.Args)
	}
	return nil
}

func SubscribeChannel(conn net.Conn, channel int) error {
	if channel < 0 {
		return fmt.Errorf("Channel number must be greater than 0")
	}

	fmt.Fprintf(conn, "subscribe channel=%d\n", channel)
	action, err := WaitForResponse(conn)
	if err != nil {
		return err
	}

	if action.Id != OK {
		return fmt.Errorf("Error expected OK, got %d, with args %v", action.Id, action.Args)
	}
	return nil
}

func HearingChannel(conn net.Conn) error {
	action, err := WaitForResponse(conn)
	if err != nil {
		return err
	}
	if action.Id != INFO {
		return fmt.Errorf("Error expected INFO, got %d, with args %v", action.Id, action.Args)
	}

	fileName := action.Args["fileName"]
	if fileName == "" {
		return fmt.Errorf("Info Header missing: %s", fileName)
	}
	file, err := os.Create(fileName)

	if err != nil {
		fmt.Printf("OS. Create() function execution error, error is:% v \n", err)
		return err
	}
	defer file.Close()
	err = CopyContent(file, conn)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	return nil
}

func Publish(conn net.Conn, channel int, message string) {
	fmt.Fprintf(conn, "publish channel=%d %s\n", channel, message)
}

func WaitForResponse(conn net.Conn) (*Action, error) {
	fmt.Println("waiting for response")
	buf := make([]byte, 50)

	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	responseMsg := buf[:n]
	action, err := NewAction(responseMsg)
	if err != nil {
		return nil, err
	}
	fmt.Printf("response: %d\n", action.Id)
	if action.Id == ERR {
		return action, fmt.Errorf("Error Response")
	} else {
		return action, nil
	}

}

func SendSuccesful(conn net.Conn) error {
	buf := make([]byte, 1024)
	okCmd := "OK\n"
	copy(buf, []byte(okCmd))
	fmt.Printf("Send: %v", buf)
	_, err := conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func SendError(conn net.Conn, errorToSend error) error {
	buf := make([]byte, 1024)
	errorMsg := fmt.Sprintf("ERR msg=%s\n", errorToSend.Error())
	copy(buf, []byte(errorMsg))
	_, err := conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
