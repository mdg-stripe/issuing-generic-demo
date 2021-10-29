package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/issuing/card"
)

func merchantHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	// get the form parameters
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	email := r.FormValue("inputEmailAddress")
	if email == "" {
		log.Fatalf("ERROR: inputEmailAddress item is empty.")
	}
	merchant := ""
	if strings.Contains(email, "fancybike") {
		merchant = "FancyBike"
	} else if strings.Contains(email, "traintix") {
		merchant = "TrainTix"
	} else if strings.Contains(email, "motoco") {
		merchant = "MotoCo"
	}

	// list all the cards we've created, filtering by merchant.
	// in practice you'd use a database to handle the mapping of merchant:card.
	// for demo purposes, we're just using the card metadata as that database mapping.
	type Card struct {
		CapturableAmount string
		CardNumber       string
		CardCVC          string
		CardExp          string
		ZipCode          string
		CardID           string
		Item             string
		CustomerName     string
	}
	var cards []Card
	params := &stripe.IssuingCardListParams{}
	params.Filters.AddFilter("limit", "", "100")
	i := card.List(params)
	for i.Next() {
		c := i.IssuingCard()
		if c.Metadata["merchant"] == merchant {
			cardParams := &stripe.IssuingCardParams{}
			cardParams.AddExpand("number")
			cardParams.AddExpand("cvc")
			pan, err := card.Get(c.ID, cardParams)
			if err != nil {
				log.Fatal(err)
			}
			cards = append(cards, Card{CapturableAmount: fmt.Sprintf("%.2f", float64(pan.SpendingControls.SpendingLimits[0].Amount)/100.0),
				CardNumber:   pan.Number,
				CardCVC:      pan.CVC,
				CardExp:      fmt.Sprintf("%d/%d", pan.ExpMonth, pan.ExpYear),
				ZipCode:      pan.Cardholder.Billing.Address.PostalCode,
				CardID:       pan.ID,
				Item:         pan.Metadata["item"],
				CustomerName: pan.Cardholder.Name})
		}
	}

	// load the HTML template
	files := []string{"./templates/merchant.page.tmpl", "./templates/base.layout.tmpl"}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	// variables for populating the template
	type Merchant struct {
		C               Config
		MerchantName    string
		Cards           []Card
		IsMerchantLogin bool
		IsIndex         bool
	}

	// render the HTML template
	p := Merchant{C: config, MerchantName: merchant,
		Cards: cards, IsMerchantLogin: true}
	err = t.Execute(w, p)
	if err != nil {
		log.Fatal(err)
	}
}
