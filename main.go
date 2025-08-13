package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ddymko/go-jsonerror"
	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"
)

const ivrMessage string = "To talk to sales, press 1. For our hours of operation, press 2. For our address, press 3."

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
	twiml, err := twiml.Voice([]twiml.Element{
		&twiml.VoiceGather{
			NumDigits: "1",
			Action:    "/gather",
			InnerElements: []twiml.Element{
				&twiml.VoiceSay{
					Message: ivrMessage,
				},
			},
		},
		&twiml.VoiceSay{
			Message: "We didn't receive any input. Goodbye!",
		},
	})
	if err != nil {
		appError(w, fmt.Errorf("could not prepare TwiML. reason: %s", err))
	}
	w.Header().Add("Content-Type", "application/xml")
	w.Write([]byte(twiml))
}

func sendSMS(recipientPhoneNumber string, twilioPhoneNumber string) (*api.ApiV2010Message, error) {
	client := twilio.NewRestClient()

	params := &api.CreateMessageParams{}
	params.SetBody("Here is our address: 375 Beale St #300, San Francisco, CA 94105, USA")
	params.SetFrom(twilioPhoneNumber)
	params.SetTo(recipientPhoneNumber)

	return client.Api.CreateMessage(params)
}

// gatherUserInput responds to the user's input
func gatherUserInput(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/xml")

	digit, err := strconv.Atoi(r.FormValue("Digits"))
	if err != nil {
		redirect := &twiml.VoiceRedirect{
			Url: "/",
		}
		say := &twiml.VoiceSay{Message: "No choice was provided. " + ivrMessage}
		twiml, err := twiml.Voice([]twiml.Element{say, redirect})
		if err != nil {
			appError(w, fmt.Errorf("could not prepare redirect TwiML. reason: %s", err))
		}
		w.Write([]byte(twiml))

		return
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

		// Send the SMS to the user
		resp, err := sendSMS(r.FormValue("From"), os.Getenv("TWILIO_PHONE_NUMBER"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			if resp.Body != nil {
				fmt.Println(*resp.Body)
			} else {
				fmt.Println(resp.Body)
			}
		}

		twimlElements = append(twimlElements, say)
	default:
		say := &twiml.VoiceSay{Message: "Sorry, I don't understand that choice."}
		twimlElements = append(twimlElements, say)
	}

	twiml, err := twiml.Voice(twimlElements)
	if err != nil {
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
	mux.HandleFunc("POST /gather", gatherUserInput)

	log.Print("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
