package AuthWebhook

import (
	"log"
	"net/http"
	"testing"
)

func TestHandleWebhook(t *testing.T) {
	// Set your secret key. Remember to switch to your live secret key in production!
	// See your keys here: https://dashboard.stripe.com/apikeys

	http.HandleFunc("/webhook", HandleWebhook)
	addr := "localhost:4242"

	log.Printf("Listening on %s ...", addr)
	t.Fatal(http.ListenAndServe(addr, nil))
}
