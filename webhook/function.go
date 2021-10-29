package AuthWebhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/issuing/authorization"
	"github.com/stripe/stripe-go/v72/webhook"
)

func HandleWebhook(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	stripe.Key = os.Getenv(("STRIPE_KEY"))
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Uncomment and replace with a real secret. You can find your endpoint's
	// secret in your webhook settings.
	//webhookSecret := "whsec_yHJJjRraIjX5HcmNjf7r8ppthog08QlJ"
	webhookSecret := os.Getenv("WEBHOOK_SECRET")

	// Verify webhook signature and extract the event.
	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), webhookSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature.
		return
	}

	if event.Type == "issuing_authorization.request" {
		var auth stripe.IssuingAuthorization
		err := json.Unmarshal(event.Data.Raw, &auth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("CARD AUTH RECEIVED\n\tName: %s\n\tAmount: %s\n\tMerchant: %s\n\tItem: %s\n\tCard last 4: %s",
			auth.Card.Metadata["name"],
			fmt.Sprintf("$%.2f", float64(auth.Card.SpendingControls.SpendingLimits[0].Amount)/100.0),
			auth.MerchantData.Name,
			auth.Card.Metadata["item"],
			auth.Card.Last4)
		handleAuthorizationRequest(auth)

		// save the auth event on Datastore so you can access it from your desktop
		dsClient, err := datastore.NewClient(ctx, os.Getenv("PROJECT_ID"))
		if err != nil {
			log.Fatal(err)
		}
		defer dsClient.Close()

		type Entity struct {
			Value []byte `datastore:",noindex"`
		}
		k := datastore.NameKey("Auths", "LastAuth", nil)
		e := new(Entity)
		e.Value = event.Data.Raw
		if _, err = dsClient.Put(ctx, k, e); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("INFO: event.Type is: %s", event.Type)
	}
	w.WriteHeader(http.StatusOK)
}

func handleAuthorizationRequest(auth stripe.IssuingAuthorization) {
	// Authorize the transaction.
	_, _ = authorization.Approve(auth.ID, &stripe.IssuingAuthorizationApproveParams{})
}
