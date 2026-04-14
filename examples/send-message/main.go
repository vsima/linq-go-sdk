// Command send-message is a runnable example that creates a chat and sends
// one text message via the Linq Partner API.
//
// Usage:
//
//	LINQ_TOKEN=xxx go run ./examples/send-message \
//	    -from +15551234567 -to +15557654321 -text "Hello from Go"
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	linq "github.com/vsima/linq-go-sdk"
)

func main() {
	from := flag.String("from", "", "sender phone in E.164 format (e.g. +15551234567)")
	to := flag.String("to", "", "recipient handle (phone or email)")
	text := flag.String("text", "", "message text")
	flag.Parse()

	if *from == "" || *to == "" || *text == "" {
		flag.Usage()
		os.Exit(2)
	}

	token := os.Getenv("LINQ_TOKEN")
	if token == "" {
		log.Fatal("LINQ_TOKEN is required")
	}

	c := linq.NewClient(token)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := c.Chats.Create(ctx, &linq.CreateChatRequest{
		From: *from,
		To:   []string{*to},
		Message: linq.CreateChatMessage{
			Parts: []linq.MessagePart{linq.NewTextPart(*text)},
		},
	})
	if err != nil {
		log.Fatalf("send: %v", err)
	}

	fmt.Printf("chat_id=%s\n", res.Chat.ID)
	if res.Chat.Message != nil {
		fmt.Printf("message_id=%s status=%s\n", res.Chat.Message.ID, res.Chat.Message.DeliveryStatus)
	}
}
