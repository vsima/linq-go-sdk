// Command webhook-server is a runnable example of a Linq webhook receiver.
//
// It verifies the HMAC-SHA256 signature on incoming requests, parses the
// event envelope, and logs a per-event summary. Point your Linq webhook
// subscription at http://<host>:8080/webhook.
//
// Env:
//
//	LINQ_SIGNING_SECRET  (required)  the secret returned when registering the subscription
//	PORT                 (optional)  listen port, default 8080
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	linq "github.com/vsima/linq-go-sdk"
)

func main() {
	secret := os.Getenv("LINQ_SIGNING_SECRET")
	if secret == "" {
		log.Fatal("LINQ_SIGNING_SECRET is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/webhook", handler(secret))
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := linq.VerifyWebhookRequest(r, secret, linq.DefaultWebhookTolerance)
		if err != nil {
			log.Printf("reject: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		evt, err := linq.ParseEvent(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("event=%s id=%s trace=%s", evt.EventType, evt.EventID, evt.TraceID)

		switch evt.EventType {
		case linq.EventMessageReceived, linq.EventMessageSent, linq.EventMessageDelivered:
			var msg linq.Message
			if err := evt.DecodeData(&msg); err == nil {
				log.Printf("  message id=%s status=%s parts=%d", msg.ID, msg.DeliveryStatus, len(msg.Parts))
			}
		case linq.EventMessageFailed:
			log.Printf("  delivery failed: %s", string(evt.Data))
		case linq.EventReactionReceived, linq.EventReactionSent:
			var r linq.Reaction
			if err := evt.DecodeData(&r); err == nil {
				log.Printf("  reaction type=%s from=%s", r.Type, r.Handle.Handle)
			}
		default:
			var generic map[string]any
			_ = json.Unmarshal(evt.Data, &generic)
			log.Printf("  data=%v", generic)
		}

		w.WriteHeader(http.StatusOK)
	}
}
