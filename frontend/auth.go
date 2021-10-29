package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/issuing/card"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
)

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// get the form parameters
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	cardID := r.FormValue("cardID")
	if cardID == "" {
		log.Fatalf("ERROR: cardID item is empty.")
	}

	// to test creating an auth on an issued card, we're following the steps here:
	// https://stripe.com/docs/issuing/testing?testing-method=with-code#retrieve-card-details
	params := &stripe.IssuingCardParams{}
	params.AddExpand("number")
	params.AddExpand("cvc")
	c, err := card.Get(cardID, params)
	if err != nil {
		log.Fatal(err)
	}

	// create a customer
	custParams := &stripe.CustomerParams{
		Address: &stripe.AddressParams{
			Country:    stripe.String("US"),
			Line1:      stripe.String("123 Main Street"),
			City:       stripe.String("San Francisco"),
			PostalCode: stripe.String("94111"),
			State:      stripe.String("CA"),
		},
		Name:        stripe.String("Jenny Rosen"),
		Email:       stripe.String("jenny.rosen@example.com"),
		Phone:       stripe.String("+18008675309"),
		Description: stripe.String("Issuing Cardholder"),
	}
	cus, err := customer.New(custParams)
	if err != nil {
		log.Fatal(err)
	}

	// create a payment method
	pmParams := &stripe.PaymentMethodParams{
		Card: &stripe.PaymentMethodCardParams{
			Number:   stripe.String(c.Number),
			ExpMonth: stripe.String(strconv.Itoa(int(c.ExpMonth))),
			ExpYear:  stripe.String(strconv.Itoa(int(c.ExpYear))),
		},
		Type: stripe.String("card"),
	}
	pm, err := paymentmethod.New(pmParams)
	if err != nil {
		log.Fatal(err)
	}

	// create an uncaptured PaymentIntent, which results in an Auth on the card
	piParams := &stripe.PaymentIntentParams{
		PaymentMethod: stripe.String(pm.ID),
		Amount:        stripe.Int64(c.SpendingControls.SpendingLimits[0].Amount),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		CaptureMethod:    stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
		Customer:         stripe.String(cus.ID),
		Confirm:          stripe.Bool(true),
	}
	piParams.AddMetadata("item", c.Metadata["item"])
	piParams.AddMetadata("name", c.Cardholder.Name)
	_, err = paymentintent.New(piParams)
	if err != nil {
		log.Fatal(err)
	}

	// sleep for 6 seconds while Stripe reaches out to our running Cloud Function in Google Cloud
	time.Sleep(6 * time.Second)

	// retrieve the auth event object, which the Google Cloud Function saved in Google Cloud Datastore
	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, config.GoogleCloudProjectID)
	if err != nil {
		log.Fatal(err)
	}
	defer dsClient.Close()

	type Entity struct {
		Value []byte `datastore:",noindex"`
	}
	k := datastore.NameKey("Auths", "LastAuth", nil)
	e := new(Entity)
	if err = dsClient.Get(ctx, k, e); err != nil {
		log.Fatal(err)
	}

	var auth stripe.IssuingAuthorization
	err = json.Unmarshal(e.Value, &auth)
	if err != nil {
		log.Fatal(err)
	}

	// load the HTML template
	files := []string{"./templates/auth.page.tmpl", "./templates/base.layout.tmpl"}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	// variables for populating the template
	type AuthInfo struct {
		C               Config
		Name            string
		Amount          string
		Merchant        string
		Item            string
		CardLast4       string
		EntireObject    string
		IsIndex         bool
		IsMerchantLogin bool
	}

	// render the HTML template
	p := AuthInfo{C: config,
		Name:            auth.Card.Cardholder.Name,
		IsMerchantLogin: true,
		Amount:          fmt.Sprintf("%.2f", float64(auth.Card.SpendingControls.SpendingLimits[0].Amount)/100.0),
		Merchant:        auth.Card.Metadata["merchant"],
		Item:            auth.Card.Metadata["item"],
		CardLast4:       auth.Card.Last4,
		EntireObject:    string(e.Value)}
	err = t.Execute(w, p)
	if err != nil {
		log.Fatal(err)
	}

}
