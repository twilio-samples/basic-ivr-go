package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"
)

const (
	// voiceName is the text-to-speech voice used throughout the IVR
	voiceName = "Google.en-US-Chirp3-HD-Aoede"
)

var requestValidator client.RequestValidator

func appError(w http.ResponseWriter, err error) {
	log.Printf("Error: %v", err)
	http.Error(w, "An error occurred processing your request", http.StatusBadRequest)
}

func validateRequest(r *http.Request) bool {
	signature := r.Header.Get("X-Twilio-Signature")
	if signature == "" {
		return false
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	url := "https://" + host + r.URL.RequestURI()

	if err := r.ParseForm(); err != nil {
		return false
	}

	params := make(map[string]string)
	for k, v := range r.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	return requestValidator.Validate(url, params, signature)
}

func twilioWebhookMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateRequest(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// handlePhoneCall receives the initial call and provides the options that the IVR supports
func handlePhoneCall(w http.ResponseWriter, r *http.Request) {
	say := &twiml.VoiceSay{
		Message: "To talk to sales, press 1. For our hours of operation, press 2. For our address, press 3.",
		Voice:   voiceName,
	}
	gather := &twiml.VoiceGather{
		NumDigits:     "1",
		Action:        "/gather",
		InnerElements: []twiml.Element{say},
	}
	twiml, err := twiml.Voice([]twiml.Element{gather})
	if err != nil {
		appError(w, fmt.Errorf("could not prepare TwiML. reason: %s", err))
		return
	}
	w.Header().Add("Content-Type", "application/xml")
	w.Write([]byte(twiml))
}

// sendAddressSMS sends an SMS with the business address
func sendAddressSMS(toPhoneNumber string) error {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromPhoneNumber := os.Getenv("TWILIO_PHONE_NUMBER")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(toPhoneNumber)
	params.SetFrom(fromPhoneNumber)
	params.SetBody("Here is our address: 8 Rue du Nom Fictif, 341, Paris")

	_, err := client.Api.CreateMessage(params)
	return err
}

// gatherUserInput responds to the user's input
func gatherUserInput(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/xml")

	digit, err := strconv.Atoi(r.FormValue("Digits"))
	if err != nil {
		redirect := &twiml.VoiceRedirect{
			Url: "/",
		}
		twiml, err := twiml.Voice([]twiml.Element{redirect})
		if err != nil {
			appError(w, fmt.Errorf("could not prepare redirect TwiML. reason: %s", err))
			return
		}
		w.Write([]byte(twiml))
		return
	}

	var twimlElements []twiml.Element
	switch digit {
	case 1:
		say := &twiml.VoiceSay{
			Message: "You selected sales. You will now be forwarded to our sales department.",
			Voice:   voiceName,
		}
		hangup := &twiml.VoiceHangup{}
		twimlElements = append(twimlElements, say, hangup)
	case 2:
		say := &twiml.VoiceSay{
			Message: "We are open from 9am to 5pm every day but Sunday.",
			Voice:   voiceName,
		}
		hangup := &twiml.VoiceHangup{}
		twimlElements = append(twimlElements, say, hangup)
	case 3:
		callerNumber := r.FormValue("From")
		if err := sendAddressSMS(callerNumber); err != nil {
			log.Printf("Error sending SMS: %s", err)
		} else {
			say := &twiml.VoiceSay{
				Message: "We will send you a text message with our address in a minute.",
				Voice:   voiceName,
			}
			hangup := &twiml.VoiceHangup{}
			twimlElements = append(twimlElements, say, hangup)
		}
	default:
		say := &twiml.VoiceSay{
			Message: "Sorry, I don't understand that choice.",
			Voice:   voiceName,
		}
		redirect := &twiml.VoiceRedirect{Url: "/"}
		twimlElements = append(twimlElements, say, redirect)
	}
	twiml, err := twiml.Voice(twimlElements)
	if err != nil {
		appError(w, fmt.Errorf("could not prepare TwiML. reason: %s", err))
		return
	}

	w.Write([]byte(twiml))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("TWILIO_AUTH_TOKEN environment variable is required")
	}
	requestValidator = client.NewRequestValidator(authToken)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", twilioWebhookMiddleware(handlePhoneCall))
	mux.HandleFunc("POST /gather", twilioWebhookMiddleware(gatherUserInput))

	log.Print("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
