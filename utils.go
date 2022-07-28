package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func CopyContent(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func ClearTerminal() {
	clear := make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear") //Darwin example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		fmt.Println("-----------------")
	}
}
func GetActionId(action string) (ACTIONID, error) {
	action = strings.ToUpper(action)
	switch action {
	case "REG":
		return REG, nil
	case "OUT":
		return OUT, nil
	case "PUB":
		return PUB, nil
	case "FILE":
		return FILE, nil
	case "SUB":
		return SUB, nil
	case "UNSUB":
		return UNSUB, nil
	case "OK":
		return OK, nil
	case "ERR":
		return ERR, nil
	default:
		return ERR, errors.New("unknown actions")
	}
}

func NewAction(message []byte) (*Action, error) {
	cmd := bytes.ToLower(bytes.TrimSpace(bytes.Split(message, []byte(" "))[0]))
	args := make(map[string]string)
	for _, arg := range bytes.Split(message, []byte(" "))[1:] {
		if bytes.Contains(arg, []byte("=")) {
			key := bytes.Split(arg, []byte("="))[0]
			value := bytes.Split(arg, []byte("="))[1]
			value = bytes.TrimSpace(value)
			args[string(key)] = string(value)
		} else {
			args[string(arg)] = ""
		}
	}
	// separe message by the first space
	fmt.Printf("cmd:%s\n", cmd)
	actionId, err := GetActionId(string(cmd))
	if err != nil {
		return nil, err
	}
	var payload []byte
	if actionId == FILE {
		content := bytes.Split(message, []byte(" "))[1:]
		payload = bytes.Join(content, []byte(" "))
	} else {
		payload = nil
	}
	return &Action{
		Id:      actionId,
		Args:    args,
		payload: payload,
	}, err

}

func GetActionText(action ACTIONID) string {
	switch action {
	case REG:
		return "register"
	case OUT:
		return "out"
	case PUB:
		return "publish"
	case FILE:
		return "file"
	case SUB:
		return "subscribe"
	case UNSUB:
		return "unsubscribe"
	case OK:
		return "ok"
	case ERR:
		return "error"
	default:
		return "unknown action"
	}
}
