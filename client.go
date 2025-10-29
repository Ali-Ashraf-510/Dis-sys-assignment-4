package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

// Message must match the server's Message struct
type Message struct {
	User string
	Text string
	Time string
}

func dialWithRetry(addr string) *rpc.Client {
	for {
		client, err := rpc.Dial("tcp", addr)
		if err == nil {
			return client
		}
		log.Printf("could not connect to server (%s): %v — retrying in 2s", addr, err)
		time.Sleep(2 * time.Second)
	}
}

func main() {
	const addr = "localhost:1234"
	fmt.Println("Simple Chatroom (type 'exit' to quit)")

	// Ask username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	nameRaw, _ := reader.ReadString('\n')
	name := strings.TrimSpace(nameRaw)
	if name == "" {
		name = "anonymous"
	}

	client := dialWithRetry(addr)
	defer client.Close()

	for {
		fmt.Print("> ")
		textRaw, err := reader.ReadString('\n') // reads whole line including spaces
		if err != nil {
			log.Printf("read error: %v", err)
			break
		}
		text := strings.TrimSpace(textRaw)
		if text == "" {
			continue
		}
		if text == "exit" {
			fmt.Println("Exiting chat.")
			break
		}

		msg := Message{
			User: name,
			Text: text,
			// Time will be assigned by server
			Time: "",
		}

		var history []Message
		callErr := client.Call("ChatServer.SendMessage", msg, &history)
		if callErr != nil {
			log.Printf("RPC error: %v", callErr)
			// try to reconnect and resend once
			client = dialWithRetry(addr)
			var retryErr error
			retryErr = client.Call("ChatServer.SendMessage", msg, &history)
			if retryErr != nil {
				log.Printf("Retry failed: %v", retryErr)
				// continue the loop so user can try again
				continue
			}
		}

		// Print full history
		fmt.Println("---- Chat History ----")
		for i, m := range history {
			fmt.Printf("%d) [%s] %s: %s\n", i+1, m.Time, m.User, m.Text)
		}
		fmt.Println("----------------------")
	}
}
