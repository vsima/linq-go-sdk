```
  ██╗     ██╗███╗   ██╗ ██████╗      ██████╗  ██████╗     ███████╗██████╗ ██╗  ██╗
  ██║     ██║████╗  ██║██╔═══██╗    ██╔════╝ ██╔═══██╗    ██╔════╝██╔══██╗██║ ██╔╝
  ██║     ██║██╔██╗ ██║██║   ██║    ██║  ███╗██║   ██║    ███████╗██║  ██║█████╔╝
  ██║     ██║██║╚██╗██║██║▄▄ ██║    ██║   ██║██║   ██║    ╚════██║██║  ██║██╔═██╗
  ███████╗██║██║ ╚████║╚██████╔╝    ╚██████╔╝╚██████╔╝    ███████║██████╔╝██║  ██╗
  ╚══════╝╚═╝╚═╝  ╚═══╝ ╚══▀▀═╝      ╚═════╝  ╚═════╝     ╚══════╝╚═════╝ ╚═╝  ╚═╝
                        iMessage · RCS · SMS — one Go client
```

[![Go Reference](https://pkg.go.dev/badge/github.com/vsima/linq-go-sdk.svg)](https://pkg.go.dev/github.com/vsima/linq-go-sdk)
[![CI](https://github.com/vsima/linq-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/vsima/linq-go-sdk/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vsima/linq-go-sdk)](https://goreportcard.com/report/github.com/vsima/linq-go-sdk)
[![codecov](https://codecov.io/gh/vsima/linq-go-sdk/branch/main/graph/badge.svg)](https://codecov.io/gh/vsima/linq-go-sdk)
[![Release](https://img.shields.io/github/v/release/vsima/linq-go-sdk?include_prereleases&sort=semver&logo=go&logoColor=white)](https://github.com/vsima/linq-go-sdk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vsima/linq-go-sdk?logo=go&logoColor=white)](./go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/vsima/linq-go-sdk/badge)](https://securityscorecards.dev/viewer/?uri=github.com/vsima/linq-go-sdk)
[![Last commit](https://img.shields.io/github/last-commit/vsima/linq-go-sdk)](https://github.com/vsima/linq-go-sdk/commits/main)
[![GitHub stars](https://img.shields.io/github/stars/vsima/linq-go-sdk?style=social)](https://github.com/vsima/linq-go-sdk/stargazers)

An unofficial Go client for the [Linq Partner API V3](https://apidocs.linqapp.com/) — send and receive iMessage, RCS, and SMS from Go.

## Install

```sh
go get github.com/vsima/linq-go-sdk@latest
```

Requires Go 1.22+. Zero third-party dependencies.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	linq "github.com/vsima/linq-go-sdk"
)

func main() {
	c := linq.NewClient(os.Getenv("LINQ_TOKEN"))
	ctx := context.Background()

	res, err := c.Chats.Create(ctx, &linq.CreateChatRequest{
		From: "+15551234567",
		To:   []string{"+15557654321"},
		Message: linq.CreateChatMessage{
			Parts: []linq.MessagePart{linq.NewTextPart("Hello from Go! 👋")},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chat:", res.Chat.ID, "message:", res.Chat.Message.ID)
}
```

## Features

- **Full V3 coverage** — chats, messages, reactions, attachments, phone numbers, webhooks
- **Typed message parts** — text / media / link with a discriminated union
- **Message effects** — screen (confetti, fireworks…) and bubble (slam, loud…) effects
- **Threaded replies** and custom-emoji reactions
- **Presigned attachment uploads** with a helper `Upload` method
- **Webhook event parsing + signature verification** (HMAC-SHA256, replay-safe)
- **Typed errors** — `APIError` with `IsNotFound`, `IsUnauthorized`, `IsRateLimited` helpers
- **Context-aware** — every call takes a `context.Context`
- **Pluggable `http.Client`** for retries, instrumentation, proxying
- **No third-party dependencies**

## Runnable examples

Two ready-to-run example programs live under [`examples/`](./examples):

```sh
# Send a single text message
LINQ_TOKEN=xxx go run ./examples/send-message \
    -from +15551234567 -to +15557654321 -text "Hello from Go"

# Receive webhooks (verifies HMAC signatures, logs each event)
LINQ_SIGNING_SECRET=xxx go run ./examples/webhook-server
```

## Snippets

### Send a message with a screen effect

```go
msg, err := c.Messages.Send(ctx, chatID, &linq.SendMessageRequest{
	Parts:  []linq.MessagePart{linq.NewTextPart("🎉")},
	Effect: &linq.MessageEffect{Type: "screen", Name: "confetti"},
})
```

### Reply in a thread

```go
msg, err := c.Messages.Send(ctx, chatID, &linq.SendMessageRequest{
	Parts:   []linq.MessagePart{linq.NewTextPart("same")},
	ReplyTo: &linq.ReplyTo{MessageID: parentID},
})
```

### React with a custom emoji

```go
emoji := "🔥"
err := c.Reactions.Set(ctx, messageID, &linq.ReactionRequest{
	Type:        linq.ReactionCustom,
	CustomEmoji: &emoji,
})
```

### Upload and send an attachment

```go
f, _ := os.Open("cat.png")
defer f.Close()

att, err := c.Attachments.Create(ctx, &linq.CreateAttachmentRequest{
	FileName: "cat.png",
	MimeType: "image/png",
})
if err != nil { log.Fatal(err) }

if err := c.Attachments.Upload(ctx, att, "image/png", f); err != nil {
	log.Fatal(err)
}

_, err = c.Messages.Send(ctx, chatID, &linq.SendMessageRequest{
	Parts: []linq.MessagePart{linq.NewMediaPartByID(att.AttachmentID)},
})
```

### Handle a signed webhook

`VerifyWebhookRequest` verifies the HMAC-SHA256 signature Linq sets on
`X-Webhook-Signature` over `{timestamp}.{rawBody}`, rejects replays older
than the tolerance, and returns the raw body for parsing.

```go
http.HandleFunc("/linq", func(w http.ResponseWriter, r *http.Request) {
	body, err := linq.VerifyWebhookRequest(r, signingSecret, linq.DefaultWebhookTolerance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	evt, err := linq.ParseEvent(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch evt.EventType {
	case linq.EventMessageReceived:
		var msg linq.Message
		_ = evt.DecodeData(&msg)
		log.Println("received:", msg.ID)
	}
	w.WriteHeader(http.StatusOK)
})
```

## Error handling

```go
_, err := c.Chats.Get(ctx, chatID)
switch {
case linq.IsNotFound(err):
	// 404
case linq.IsUnauthorized(err):
	// 401
case linq.IsRateLimited(err):
	var ae *linq.APIError
	errors.As(err, &ae)
	time.Sleep(time.Duration(ae.RetryAfter) * time.Second)
case err != nil:
	log.Fatal(err)
}
```

## Testing

```sh
make test           # go test ./...
make cover          # coverage report
make lint           # gofmt + go vet
```

## Contributing

Issues and PRs welcome. See [CONTRIBUTING.md](./CONTRIBUTING.md).

## Disclaimer

This SDK is **not affiliated with Linq**. Product names and trademarks are property of their respective owners.

## License

[MIT](./LICENSE) © Victor Sima
