# Stripe Issuing - Generic Demo

This demo shows how to:
* create virtual cards with Stripe Issuing; 
* set spending controls; 
* expose raw PAN's from Stripe-issued cards;
* handle real-time auth's using a live webhook in Google Cloud Functions

Use cases could include travel agents (who collect payments in one lump, and then book with multiple vendors), buy-now-pay-later providers (who finance a purchase and manage installment collection), and others who need to issue cards.

The demo backend is written in the Go programming language; the frontend is mostly plain HTML styled with Bootstrap.

## Setup
0. Install prerequisite software:
* Go programming language: https://golang.org/dl/
* Google Cloud SDK: cloud.google.com/sdk

1. Create a test account for Stripe Issuing by following the scenario [here](https://admin.corp.stripe.com/scenarios/team/issuing). Give it a generous issuing top-up balance (default is $2500). Log in and note the Stripe secret key, which you'll need for the steps below.

2. While logged in with your @stripe.com account, navigate to console.cloud.google.com. 
  * Create a new project if you don't already have one.
  * You may have to enter your own credit card to enable the account (try this guide with GCP's "free mode" first). Although this demo uses virtually no paid resources, set a reminder to delete the GCP account at a future date, so you're not charged for anything.
  * Navigate to the "Cloud Functions" and "Cloud Datastore" sections of the GCP console to enable those services.

3. Create a GCP Service Account, grant it the Owner IAM role, and download it to your laptop.
  * From the GCP console: Menu bar in top left -> IAM & Admin -> Service Accounts. 
  * Name the service account whatever you want. 
  * Under "Grant this service account access to a project," choose Select a role -> Basic -> Owner.
  * Skip "Grant users access to this service account."
  * Once the service account has been created, click the three dots to the right, and select "Manage keys." 
  * In the new dialog, click Add Key -> Create new key -> JSON -> Create. This downloads a .json file to your laptop. **Do not give this .json file to anyone else, as they can do anything in your GCP project with it**. 

4. Update `issuing-generic-demo/webhook/env.yaml`:
* Go to https://dashboard.stripe.com/settings/issuing/authorizations and enter `https://www.test.com` as an authorization webhook (you'll update this later). Click save.
* Then go to https://dashboard.stripe.com/test/webhooks and click on the test.com webhook. Under "Signing secret," click Reveal. 
* Copy this value to `WEBHOOK_SECRET` in  `issuing-generic-demo/webhook/env.yaml`.
* Also update `env.yaml` to include your Stripe secret key, and your Google Cloud project ID (find this in the GCP Console -> top left menu -> Home -> note the project ID in the top left (note: you want project ID, not project name!))
* Also put this GCP project ID in `PROJECT_ID` in `webhook/deploy_function.sh`.

5. Open a Terminal prompt. `cd` to the root of this demo (if you type `ls` you should see `frontend`, `webhook`, and `README.md`. Type:
```
export GOOGLE_APPLICATION_CREDENTIALS=<absolute path to your .json service account key>
```
  * You will have to do this every time you launch a new Terminal. Consider [adding this value to your PATH](https://www.cyberciti.biz/faq/appleosx-bash-unix-change-set-path-environment-variable/).

6. `cd` to `webhook`. Type `go build`.  If you get error messages about importing packages, type `go mod tidy`. Then try `go build` again. If it returns without printing anything to the screen, that means it succeeded.
  * If the `go` command gives you `command not found`, [edit your PATH](https://www.cyberciti.biz/faq/appleosx-bash-unix-change-set-path-environment-variable/) to include /usr/local/go/bin

7. In `webhook`, type `chmod +x deploy_function.sh`. Then run the script by typing `./deploy_function.sh`.
  * If there are errors, please try and solve them, else email mdg@stripe.com. 

8. In the output from the last step, look for the `httpsTrigger` URL value (should end with `/RealTimeWebhook`).
  * Copy the entire URL.
  * Go to the Stripe dashboard -> search at the top for Webhooks -> click on your test.com webhook.  
  * Click on the `...` button in the top right, select "Update details," and paste the URL of the live webhook.
  * This webhook is now ready to receive real-time auths from Stripe. It will error out with HTTP 400 if you click it though, since it expects a webhook secret in every request.

9. (Optional) Save the customer's logo to `issuing-generic-demo/frontend/templates/img`. 

10. Edit `issuing-generic-demo/frontend/config.json`. 
  * Most things should be self-explanatory.
  * `brand_color_html_hex` will be your navbar color.
  * `navbar-light` means black navbar text, and `navbar-dark` means white navbar text (yes, you read that correctly).
  * `merchant_logo` is simply the base name of the file. You should NOT include any file paths.

11. In your terminal `cd` to `issuing-generic-demo/frontend`. Type `go build`. If you get error messages about importing packages, type `go mod tidy`. Then try `go build` again. If it returns without printing anything to the screen, that means it succeeded.

12. Still in `frontend`, type:
```
go run main.go auth.go merchant.go payprovider.go
```
  * If you're on Mac OS X, a popup should ask you to listen on the port you specified in config.json. Click Allow.
  * Open a web browser and go to: http://localhost:8080/ (or whatever port number you put in config.json)

13. In the "Logged in user" box, type any text at all like "John Doe" and click "Pay with MagicPay." You should immediately see a receipt.
  * In a separate tab, go to https://dashboard.stripe.com/test/issuing/overview and verify that a card was issued to John Doe, with the purchase amount as a spending control.

14. Go back to the tab with the localhost demo. Click on "Merchant Experience." Enter "test@traintix.com" and anything for the password field. You should see a list of raw PAN's that the merchant can swipe manually or programmatically. 

15. Click Simulate Auth **only once** (very important). The webpage will hang for 6-10 seconds while waiting for Stripe to invoke our webhook in GCP. The GCP webhook then stores the auth object in Google Cloud Datastore; and finally, our local demo retrieves that auth object and displays it on the next page.
  * Watch the Terminal prompt at this point. If it crashes and the program exits, start over from step 12. (The program crashes if the Cloud Function doesn't respond within 2 seconds, which sometimes happens).


## References
* [Creating cards](https://stripe.com/docs/issuing/cards)
* [Info about PCI](https://stripe.com/docs/issuing/cards/virtual#details-about-pci-dss) - your customer needs to read this if they're planning on handing raw PAN's to merchants
* [Spending controls](https://stripe.com/docs/issuing/controls/spending-controls)
* [Real-time auths](https://stripe.com/docs/issuing/controls/real-time-authorizations)
* [Disputes](https://stripe.com/docs/issuing/purchases/disputes)
* [Stripe Go client library documentation](https://pkg.go.dev/github.com/stripe/stripe)
