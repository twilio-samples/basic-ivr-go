package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tj/assert"
)

type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",innerxml"`
	Nodes   []node     `xml:",any"`
}

// decodeXML is a simplistic way of creating an XML object from the string content provided
func decodeXML(xmlString string) (node, error) {
	var xmlObj node
	err := xml.NewDecoder(strings.NewReader(xmlString)).Decode(&xmlObj)
	return xmlObj, err
}

func TestCanHandlePhoneCall(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Expected response",
			want: fmt.Sprintf(
				`<?xml version="1.0" encoding="UTF-8"?><Response><Gather action="/gather" numDigits="1"><Say>%s</Say></Gather><Say>We didn&apos;t receive any input. Goodbye!</Say></Response>`,
				ivrMessage,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			r, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			handlePhoneCall(rr, r)

			rs := rr.Result()
			assert.Equal(t, rs.StatusCode, http.StatusOK)

			defer rs.Body.Close()
			body, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			body = bytes.TrimSpace(body)

			expected, err := decodeXML(tt.want)
			assert.Nil(t, err)
			obtained, err := decodeXML(string(body))
			assert.Nil(t, err)
			assert.Equal(t, expected, obtained)

			assert.Equal(t, "application/xml", rs.Header.Get("Content-Type"))
		})
	}
}

func TestCanGatherUserInput(t *testing.T) {
	tests := []struct {
		name  string
		digit string
		want  string
	}{
		{
			name:  "User provided an empty digit",
			digit: "",
			want: fmt.Sprintf(
				`<?xml version="1.0" encoding="UTF-8"?><Response><Say>No choice was provided. %s</Say><Redirect>/</Redirect></Response>`,
				ivrMessage,
			),
		},
		{
			name: "User provided no digit",
			want: fmt.Sprintf(
				`<?xml version="1.0" encoding="UTF-8"?><Response><Say>No choice was provided. %s</Say><Redirect>/</Redirect></Response>`,
				ivrMessage,
			),
		},
		{
			name:  "User pressed 1",
			digit: "1",
			want:  `<?xml version="1.0" encoding="UTF-8"?><Response><Say>You selected sales. You will now be forwarded to our sales department.</Say></Response>`,
		},
		{
			name:  "User pressed 2",
			digit: "2",
			want:  `<?xml version="1.0" encoding="UTF-8"?><Response><Say>We are open from 9am to 5pm every day but Sunday.</Say></Response>`,
		},
		{
			name:  "User provided an unknown digit",
			digit: "4",
			want:  `<?xml version="1.0" encoding="UTF-8"?><Response><Say>Sorry, I don&apos;t understand that choice.</Say></Response>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize a new httptest.ResponseRecorder.
			rr := httptest.NewRecorder()

			// Initialize a new dummy http.Request.
			r, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			r.ParseForm()
			r.Form.Add("Digits", tt.digit)

			gatherUserInput(rr, r)

			rs := rr.Result()
			assert.Equal(t, rs.StatusCode, http.StatusOK)

			defer rs.Body.Close()
			body, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			body = bytes.TrimSpace(body)

			assert.Equal(t, tt.want, string(body))
		})
	}
}
