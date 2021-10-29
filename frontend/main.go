package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v72"
)

// struct and global variable for parsing the config.json file
type Config struct {
	StripeSecretKey                      string `json:"stripe_secret_key"`
	BrandColor                           string `json:"brand_color_html_hex"`
	AccentColor                          string `json:"accent_color_html_hex"`
	BootstrapNavbarTextColor             string `json:"bootstrap_navbar_text_color"` // acceptable values are 'navbar-light' (for black navbar text) and 'navbar-dark' (for white navbar text)
	HtmlTitleText                        string `json:"html_title_text"`
	MerchantLogo                         string `json:"merchant_logo"` //just the filename, not the path. The code assumes it's in templates/img/
	GoogleCloudServiceAccountKeyLocation string `json:"google_cloud_service_account_key_location"`
	GoogleCloudProjectID                 string `json:"google_cloud_project_id"`
	LocalhostPortString                  string `json:"localhost_port_string"`
}

var config Config

func main() {
	// parse the config file
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	stripe.Key = config.StripeSecretKey
	if config.StripeSecretKey == "" {
		log.Fatalf("ERROR: config.json: must specify stripe_secret_key, which is the secret key you can find in the Stripe Dashboard.")
	}

	log.Print("Starting server...")

	// debug: to see the raw, unrendered html template, uncomment the line below
	// http.Handle("/templates", http.FileServer(http.Dir("./templates/")))

	// handle requests for static files
	imgDir := http.FileServer(http.Dir("./templates/img"))
	http.Handle("/img/", http.StripPrefix("/img/", imgDir))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/payprovider", payproviderHandler)
	http.HandleFunc("/merchant", merchantHandler)
	http.HandleFunc("/auth", authHandler)

	// Determine port for HTTP service.
	port := config.LocalhostPortString
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// struct for values that will be passed to the template
	type IndexResponse struct {
		C               Config
		IsIndex         bool
		IsMerchantLogin bool
	}

	files := []string{"./templates/home.page.tmpl", "./templates/base.layout.tmpl"}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	indexVars := IndexResponse{C: config, IsIndex: true}
	err = t.Execute(w, indexVars)
	if err != nil {
		log.Fatal(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginResponse struct {
		C               Config
		IsIndex         bool
		IsMerchantLogin bool
	}

	files := []string{"./templates/login.page.tmpl", "./templates/base.layout.tmpl"}
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, LoginResponse{C: config, IsMerchantLogin: true})
	if err != nil {
		log.Fatal(err)
	}
}
