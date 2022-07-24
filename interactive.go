package main

import (
	"fmt"
	"net"
	"os"
)

func Interative(conn net.Conn) error {

	err := RegisterToServer(conn)
	if err != nil {
		return err
	}
	for {

		command, err := menu()
		if err != nil {
			return err
		}

		switch command {
		case "1":
			err = subscribe(conn)
			if err != nil {
				return err
			}
		case "2":
			err = publish(conn)
			if err != nil {
				return err
			}
		case "3":
			fmt.Printf("Exiting\n")
			os.Exit(0)
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}

}

func menu() (string, error) {
	fmt.Printf("Enter command: \n 1. Subscribe to channel \n 2. Publish to channel \n 3. Exit \n")
	var command int
	fmt.Scanf("%d", &command)
	ClearTerminal()
	if command < 1 || command > 3 {
		return "", fmt.Errorf("unknown command %d", command)
	}

	return fmt.Sprintf("%d", command), nil
}

func subscribe(conn net.Conn) error {
	fmt.Printf("Enter channel number: ")
	var channel int
	fmt.Scanf("%d", &channel)
	err := SubscribeToChannel(conn, channel)
	if err != nil {
		return err
	}
	for {

		err = HearingChannel(conn)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return err
		}
	}

}

func publish(conn net.Conn) error {
	fmt.Printf("Enter file path: ")
	var filePath string
	fmt.Scanf("%s", &filePath)
	var channelToPublish int
	fmt.Printf("Enter channel to publish: ")
	fmt.Scanf("%d", &channelToPublish)
	err := SendFile(conn, filePath, channelToPublish)
	if err != nil {
		return err
	}
	return nil
}
