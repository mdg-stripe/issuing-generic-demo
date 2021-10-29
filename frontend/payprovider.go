package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/issuing/card"
	"github.com/stripe/stripe-go/v72/issuing/cardholder"
)

func payproviderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	// get the form parameters
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	item := r.FormValue("item")
	if item == "" {
		log.Fatalf("ERROR: payproviderHandler item is empty.")
	}
	acctId := r.FormValue("acctId")
	if acctId == "" {
		log.Fatalf("ERROR: payproviderHandler acctId is empty.")
	}
	user := r.FormValue("user")
	if user == "" {
		log.Fatalf("ERROR: payproviderHandler user is empty.")
	}
	sessionRateString := r.FormValue("rate")
	if sessionRateString == "" {
		log.Fatalf("ERROR: payproviderHandler rate is empty.")
	}
	sessionRate, err := strconv.Atoi(sessionRateString)
	if err != nil {
		log.Fatal(err)
	}
	sessionRate = sessionRate * 100 // need to multiply the dollar amount by 100, since Stripe expects prices in cents instead of dollars
	log.Printf("%s, %s, %s, %d", item, acctId, user, sessionRate)

	// CREATE A CARDHOLDER AND ISSUE A CARD
	// First, create the cardholder
	holderParams := &stripe.IssuingCardholderParams{
		Billing: &stripe.IssuingCardholderBillingParams{
			Address: &stripe.AddressParams{
				Country:    stripe.String("US"),
				Line1:      stripe.String("123 Main Street"),
				City:       stripe.String("San Francisco"),
				PostalCode: stripe.String("94111"),
				State:      stripe.String("CA"),
			},
		},
		Email:       stripe.String("test@demo.com"),
		Name:        stripe.String(user),
		PhoneNumber: stripe.String("+18008675309"),
		Status:      stripe.String(string(stripe.IssuingCardholderStatusActive)),
		Type:        stripe.String(string(stripe.IssuingCardholderTypeIndividual)),
	}

	ch, err := cardholder.New(holderParams)
	if err != nil {
		log.Fatal(err)
	}

	// Next, create the card
	cardParams := &stripe.IssuingCardParams{
		Cardholder: stripe.String(ch.ID),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		Type:       stripe.String("virtual"),
		Status:     stripe.String("active"),
		SpendingControls: &stripe.IssuingCardSpendingControlsParams{
			/*AllowedCategories: []*string{
				stripe.String("car_rental_agencies"),
			},*/
			SpendingLimits: []*stripe.IssuingCardSpendingControlsSpendingLimitParams{
				{
					Amount:   stripe.Int64(int64(sessionRate)),
					Interval: stripe.String(string(stripe.IssuingCardSpendingControlsSpendingLimitIntervalAllTime)),
				},
			},
		},
	}
	// optional: tag the card with metadata key-value pairs. These have no impact on Stripe, but your application
	// logic may use these key-value pairs in whatever way you like
	cardParams.AddMetadata("item", item)
	cardParams.AddMetadata("user", user)
	cardParams.AddMetadata("merchant", acctId)
	c, err := card.New(cardParams)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", c)

	// load the HTML template
	files := []string{"./templates/payprovider.page.tmpl", "./templates/base.layout.tmpl"}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	// variables for populating the template
	type PayProvider struct {
		C               Config
		Item            string
		SessionRate     string
		Merchant        string
		User            string
		IsIndex         bool
		IsMerchantLogin bool
		CardAPIObject   string
	}

	// render the HTML template
	obj, _ := json.MarshalIndent(c, "", "    ")
	p := PayProvider{C: config, Item: item,
		SessionRate:   fmt.Sprintf("%.2f", float64(sessionRate)/100.0),
		Merchant:      acctId,
		User:          user,
		CardAPIObject: string(obj)}
	err = t.Execute(w, p)
	if err != nil {
		log.Fatal(err)
	}
}
