# AGENTS.md

## Development standards

* Follow Go standard formatting (use `go fmt`)
* Use meaningful variable and function names

## Immutable code

Don't change this code:

* Webhook signature validation logic (implemented in `twilioWebhookMiddleware` using `client.RequestValidator`)

## Important commands

* `go test` to run tests
* `go run main.go` to start the app
* `go fmt` to format code

## Common mistakes

* Don't send phone numbers in a format other than [E.164 format](https://en.wikipedia.org/wiki/E.164)

## Task cookbook

### Adding a new IVR option

1. Update the initial call handler `handlePhoneCall` to mention the new option in the `<Say>` message
2. Add a new case in the `gatherUserInput` function's switch statement
3. Include appropriate TwiML elements (`<Say>`, `<Hangup>`, etc.)
4. Add unit tests for the new option
