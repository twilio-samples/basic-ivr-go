package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/ddymko/go-jsonerror"
	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go/twiml"
)

func appError(w http.ResponseWriter, err error) {
	var error jsonerror.ErrorJSON
	error.AddError(jsonerror.ErrorComp{
		Detail: err.Error(),
		Code:   strconv.Itoa(http.StatusBadRequest),
		Title:  "Something went wrong",
		Status: http.StatusBadRequest,
	})
	http.Error(w, error.Error(), http.StatusBadRequest)
}

// handlePhoneCall receives the initial call and provides the options that the IVR supports
func handlePhoneCall(w http.ResponseWriter, r *http.Request) {
	gather := &twiml.VoiceGather{
		NumDigits: "1",
		Action:    "/gather",
	}
	say := &twiml.VoiceSay{
		Message: "To talk to sales, press 1. For our hours of operation, press 2. For our address, press 3.",
	}
	redirect := &twiml.VoiceRedirect{
		Url: "/",
	}
	twiml, err := twiml.Voice([]twiml.Element{gather, say, redirect})
	if err == nil {
		appError(w, fmt.Errorf("could not prepare TwiML. reason: %s", err))
	}
	w.Header().Add("Content-Type", "application/xml")
	w.Write([]byte(twiml))
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
		if err == nil {
			appError(w, fmt.Errorf("could not prepare redirect TwiML. reason: %s", err))
		}
		w.Write([]byte(twiml))
	}

	var twimlElements []twiml.Element
	switch digit {
	case 1:
		say := &twiml.VoiceSay{Message: "You selected sales. You will now be forwarded to our sales department."}
		twimlElements = append(twimlElements, say)
	case 2:
		say := &twiml.VoiceSay{Message: "We are open from 9am to 5pm every day but Sunday."}
		twimlElements = append(twimlElements, say)
	case 3:
		say := &twiml.VoiceSay{Message: "We will send you a text message with our address in a minute."}
		twimlElements = append(twimlElements, say)
	default:
		say := &twiml.VoiceSay{Message: "Sorry, I don't understand that choice."}
		redirect := &twiml.VoiceRedirect{Url: "/"}
		twimlElements = append(twimlElements, say, redirect)
	}
	twiml, err := twiml.Voice(twimlElements)
	if err == nil {
		appError(w, fmt.Errorf("could not prepare TwiML. reason: %s", err))
	}

	w.Write([]byte(twiml))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handlePhoneCall)
	mux.HandleFunc("POST /sms", gatherUserInput)

	log.Print("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
