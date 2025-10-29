# Basic IVR system with Go

This app shows you how to build a basic interactive voice response (IVR) system using Go and Twilio.

## IVR overview

This sample app provides the following functionality:

1. A customer calls your Twilio phone number
2. Your app answers the call using text-to-speech and asks the caller to choose from one of three options:
    1. Talk to sales 
    2. Get the company's hours of operation
    3. Get the company's address
3. The caller dials an option. Your app tells them the information that they requested or it sends an SMS to the caller with the company's address.

## Prerequisites

To run the app locally, you need the following:

- Go 1.22.5 or later
- An [ngrok](https://ngrok.com/) account
- A [Twilio account](https://www.twilio.com/try-twilio) with an active phone number that can send SMS

## Quickstart

1. Clone or download this repository.
2. Install the dependencies:
    ```bash
    go mod tidy
    ```
3. Create a `.env` file in the project root:
    ```bash
    cp .env.example .env
    ```
    Or create it manually with the following content:
    ```env
    TWILIO_ACCOUNT_SID=your_account_sid_here
    TWILIO_AUTH_TOKEN=your_auth_token_here
    TWILIO_PHONE_NUMBER=your_twilio_phone_number_here
    ```
4. Go to the [Twilio Console](https://console.twilio.com/) and find your **Account SID**, **Auth Token**, and Twilio phone number.
5. Copy and paste those values into the placeholders in the `.env` file. Save the file.
6. Start the app:
    ```bash
    go run main.go
    ```
7. Start your ngrok server:
    ```bash
    ngrok http 8080
    ```
8. Go to the [Active numbers](https://www.twilio.com/console/phone-numbers/incoming) page in the Twilio Console.
9. Click your Twilio phone number.
10. Go to the **Configure** tab and find the **Voice Configuration** section.
11. In the **A call comes in** row, select the **Webhook** option.
12. Paste your ngrok public URL in the **URL** field. For example, if your ngrok console shows Forwarding `https://1aaa-123-45-678-910.ngrok-free.app`, enter `https://1aaa-123-45-678-910.ngrok-free.app`.
13. Click **Save configuration**.
14. With the Go server and ngrok running, call your Twilio phone number. You hear the IVR greeting defined in `main.go`.