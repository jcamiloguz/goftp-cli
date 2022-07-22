package main

import (
	"fmt"
	"net"
	"os"
)

func RegisterToServer(conn net.Conn) error {
	fmt.Fprintf(conn, "register \n")
	action, err := WaitForResponse(conn)
	if err != nil {
		return err
	}
	if action.Id != OK {
		return fmt.Errorf("Error expected OK, got %d, with args %v", action.Id, action.Args)
	}
	fmt.Println("Registered to server")
	return nil
}

func SubscribeToChannel(conn net.Conn, channel int) error {
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
func SendFile(conn net.Conn, filePath string, channel int) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("OS. Stat() function execution error, error is:% v \n", err)
		return err
	}
	sizeInKb := fileInfo.Size() / 1024
	fmt.Printf("file %s %dKb\n", fileInfo.Name(), sizeInKb)

	publishMsg := fmt.Sprintf("publish channel=%d  fileName=%s size=%d\n", channel, fileInfo.Name(), fileInfo.Size())
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
