<!-- markdownlint-disable MD013 -->
# Basic IVR with Go

This app shows how to build a basic [IVR (Interactive voice response)](https://www.twilio.com/en-us/use-cases/ivr) system with Go and Twilio.

Find out more on [Twilio Code Exchange](https://www.twilio.com/code-exchange/basic-ivr).

## Application Overview

- A user calls their Twilio phone number
- The user is then presented with 3 options:
  1. Talk to sales
  2. The company's hours of operation
  3. The company's address
- If the user chooses one of the first two options, they get a voice response on the call with more information
- If they choose the third option, they will receive an SMS with the company's address information

## Requirements

To use the application, you'll need the following:

- [Go](https://go.dev/doc/install) 1.22 or above
- A Twilio account (free or paid) with a phone number. [Click here to create one](http://www.twilio.com/try-twilio), if you don't have already.
- [ngrok](https://ngrok.com/)
- A network testing tool such as [curl](https://curl.se/)
<!-- markdownlint-enable MD013 -->