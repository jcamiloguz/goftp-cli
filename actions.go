package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

const BUFFER_SIZE = 1024

func RegisterToServer(conn net.Conn) error {
	fmt.Fprintf(conn, "register \n")
	action, err := WaitForResponse(conn, nil)
	if err != nil {
		return err
	}
	if action.Id != OK {
		return fmt.Errorf("error expected OK, got %d, with args %v", action.Id, action.Args)
	}
	fmt.Println("Registered to server")
	return nil
}

func SubscribeToChannel(conn net.Conn, channel int) error {
	if channel < 0 {
		return fmt.Errorf("channel number must be greater than 0")
	}

	fmt.Fprintf(conn, "subscribe channel=%d\n", channel)
	action, err := WaitForResponse(conn, nil)
	if err != nil {
		return err
	}

	if action.Id != OK {
		return fmt.Errorf("error expected OK, got %d, with args %v", action.Id, action.Args)
	}
	return nil
}

func HearingChannel(conn net.Conn) error {
	messageToPrint := "Hearing channel"
	action, err := WaitForResponse(conn, &messageToPrint)
	if err != nil {
		return err
	}
	if action.Id != PUB {
		return fmt.Errorf("error expected PUB, got %d, with args %v", action.Id, action.Args)
	}

	fileName := action.Args["fileName"]
	size := action.Args["size"]
	if fileName == "" {
		return fmt.Errorf("info header missing: %s", fileName)
	}
	//parse int
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("info header missing: %s", size)
	}

	chunks := sizeInt / (BUFFER_SIZE - 5)
	garbage := (BUFFER_SIZE - 5) - (sizeInt % (BUFFER_SIZE - 5))
	if garbage > 0 {
		chunks++
	}
	fmt.Printf("size:%dgarbage: %d, chunks:%d\n", sizeInt, garbage, chunks)

	file, err := os.Create(fileName)

	if err != nil {
		fmt.Printf("OS. Create() function execution error, error is:% v \n", err)
		return err
	}
	defer file.Close()
	for i := 1; i < chunks+2; i++ {

		buffeFull := make([]byte, BUFFER_SIZE)
		n, err := conn.Read(buffeFull)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading file: %s\n", err)
			return err
		}
		fmt.Printf("buffer:%s\n", buffeFull[:n])

		action, err = NewAction(buffeFull[:n])
		if err != nil {
			fmt.Printf("Error creating action: %s\n", err)
			return err
		}

		switch action.Id {
		case FILE:

			fmt.Println("Writing file")

			if i >= chunks && garbage > 0 {
				fmt.Println("Cleaning last chunk")
				payload := action.payload[:BUFFER_SIZE-5-garbage]
				_, err = file.Write(payload)
				if err != nil {
					fmt.Printf("Error writing file: %s\n", err)
					return err
				}
				// fmt.Printf("writed: %s\n", string(payload))

			} else {
				_, err = file.Write(action.payload)
				if err != nil {
					fmt.Printf("Error writing file: %s\n", err)
					return err
				}
			}

		case OK:
			fmt.Println("OK")
			return nil
		default:
			fmt.Printf("Unknown action: %d\n", action.Id)

		}

	}

	return nil
}

func WaitForResponse(conn net.Conn, msg *string) (*Action, error) {
	if msg == nil {
		fmt.Println("waiting for response")
	} else {
		fmt.Printf("%s\n", *msg)
	}
	buf := make([]byte, BUFFER_SIZE)

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
		return action, fmt.Errorf("error response")
	} else {
		return action, nil
	}

}

func SendSuccesful(conn net.Conn) error {
	buf := make([]byte, BUFFER_SIZE)
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
	buf := make([]byte, BUFFER_SIZE)
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
	sizeInKb := fileInfo.Size() / BUFFER_SIZE
	fmt.Printf("file %s %dKb\n", fileInfo.Name(), sizeInKb)

	publishMsg := fmt.Sprintf("publish channel=%d  fileName=%s size=%d\n", channel, fileInfo.Name(), fileInfo.Size())
	_, err = conn.Write([]byte(publishMsg))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	// _, err = WaitForResponse(conn, nil)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// 	return err
	// }

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	for {
		buffeFull := make([]byte, BUFFER_SIZE)
		contentBuffer := make([]byte, 1019)
		n, err := file.Read(contentBuffer)
		if err != nil {
			if err == io.EOF {
				_, err := conn.Write([]byte("ok \n"))
				if err != nil {
					return err
				}
				break
			}
			fmt.Printf("Error reading file: %s\n", err)
			return err
		}
		// add FILE header to buffer
		contentHeader := "file "

		//full buffer is contentHeader + contentBuffer
		copy(buffeFull, []byte(contentHeader))
		copy(buffeFull[len(contentHeader):], contentBuffer[:n])

		_, err = conn.Write(buffeFull)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return err
		}

		// fmt.Printf("Send: %v\n", string(buffeFull))
	}
	action, err := WaitForResponse(conn, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	if action.Id != OK {
		return fmt.Errorf("error response")
	}
	fmt.Printf("File sent\n")

	return nil
}
