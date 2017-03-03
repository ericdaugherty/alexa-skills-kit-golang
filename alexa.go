package alexa

import (
	"errors"
	"log"
	"math"
	"strconv"
	"time"
)

const sdkVersion = "1.0"
const launchRequestName = "LaunchRequest"
const intentRequestName = "IntentRequest"
const sessionEndedRequestName = "SessionEndedRequest"

var timestampTolerance = 150

// Alexa defines the primary interface to use to create an Alexa request handler.
type Alexa struct {
	ApplicationID   string
	RequestHandler  RequestHandler
	IgnoreTimestamp bool
}

// RequestHandler defines the interface that must be implemented to handle
// Alexa Requests
type RequestHandler interface {
	OnSessionStarted(*Request, *Session, *Response) error
	OnLaunch(*Request, *Session, *Response) error
	OnIntent(*Request, *Session, *Response) error
	OnSessionEnded(*Request, *Session, *Response) error
}

// RequestEnvelope contains the data passed from Alexa to the request handler.
type RequestEnvelope struct {
	Version string   `json:"version"`
	Session *Session `json:"session"`
	Request *Request `json:"request"`
	// TODO Add Request Context
}

// Session containes the session data from the Alexa request.
type Session struct {
	New        bool   `json:"new"`
	SessionID  string `json:"sessionId"`
	Attributes struct {
		String map[string]interface{} `json:"string"`
	} `json:"attributes"`
	User struct {
		UserID      string `json:"userId"`
		AccessToken string `json:"accessToken"`
	} `json:"user"`
	Application struct {
		ApplicationID string `json:"applicationId"`
	} `json:"application"`
}

// Request contines the data in the request within the main request.
type Request struct {
	Locale    string `json:"locale"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Intent    Intent `json:"intent"`
	Name      string `json:"name"`
}

// Intent contains the data about the Alexa Intent requested.
type Intent struct {
	Name  string                `json:"name"`
	Slots map[string]IntentSlot `json:"slots"`
}

// IntentSlot contains the data for one Slot
type IntentSlot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ResponseEnvelope contains the Response and additional attributes.
type ResponseEnvelope struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          *Response              `json:"response"`
}

// Response contains the body of the response.
type Response struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	Directives       *[]Directive  `json:"directives,omitempty"`
	ShouldSessionEnd bool          `json:"shouldEndSession"`
}

// OutputSpeech contains the data the defines what Alexa should say to the user.
type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

// Card contains the data displayed to the user by the Alexa app.
type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

// Image provides URL(s) to the image to display in resposne to the request.
type Image struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

// Reprompt contains data about whether Alexa should prompt the user for more data.
type Reprompt struct {
	OutputSpeech *OutputSpeech `json:"outputSpeech,omitempty"`
}

// Directive contains device level instructions on how to handle the response.
type Directive struct {
	Type         string `json:"type"`
	PlayBehavior string `json:"playBehavior,omitempty"`
	AudioItem    *struct {
		Stream *Stream `json:"stream,omitempty"`
	} `json:"audioItem,omitempty"`
}

// Stream contains instructions on playing an audio stream.
type Stream struct {
	Token                string `json:"token"`
	URL                  string `json:"url"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
}

// ProcessRequest handles a request passed from Alexa
func (alexa *Alexa) ProcessRequest(requestEnv *RequestEnvelope) (*ResponseEnvelope, error) {

	err := alexa.verifyApplicationID(requestEnv)
	if err != nil {
		return nil, err
	}
	if !alexa.IgnoreTimestamp {
		err = alexa.verifyTimestamp(requestEnv)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Ignoring timestamp verification.")
	}

	request := requestEnv.Request
	session := requestEnv.Session

	responseEnv := &ResponseEnvelope{}
	responseEnv.Version = sdkVersion
	responseEnv.Response = &Response{}
	responseEnv.Response.ShouldSessionEnd = true // Set default value.

	response := responseEnv.Response

	// If it is a new session, invoke onSessionStarted
	if session.New {
		err := alexa.RequestHandler.OnSessionStarted(request, session, response)
		if err != nil {
			log.Println("Error handling OnSessionStarted.", err.Error())
			return nil, err
		}
	}

	switch requestEnv.Request.Type {
	case launchRequestName:
		err := alexa.RequestHandler.OnLaunch(request, session, response)
		if err != nil {
			log.Println("Error handling OnLaunch.", err.Error())
			return nil, err
		}
	case intentRequestName:
		err := alexa.RequestHandler.OnIntent(request, session, response)
		if err != nil {
			log.Println("Error handling OnIntent.", err.Error())
			return nil, err
		}
	case sessionEndedRequestName:
		err := alexa.RequestHandler.OnSessionEnded(request, session, response)
		if err != nil {
			log.Println("Error handling OnSessionEnded.", err.Error())
			return nil, err
		}
	}

	return responseEnv, nil
}

// SetTimestampTolerance sets the maximum number of seconds to allow between
// the current time and the request Timestamp.  Default value is 150 seconds.
func (alexa *Alexa) SetTimestampTolerance(seconds int) {
	timestampTolerance = seconds
}

// SetSimpleCard creates a new simple card with the specified content.
func (r *Response) SetSimpleCard(title string, content string) {
	r.Card = &Card{Type: "Simple", Title: title, Content: content}
}

// SetStandardCard creates a new standard card with the specified content.
func (r *Response) SetStandardCard(title string, text string, smallImageURL string, largeImageURL string) {
	r.Card = &Card{Type: "Standard", Title: title, Text: text}
	r.Card.Image = &Image{SmallImageURL: smallImageURL, LargeImageURL: largeImageURL}
}

// SetLinkAccountCard creates a new LinkAccount card.
func (r *Response) SetLinkAccountCard() {
	r.Card = &Card{Type: "LinkAccount"}
}

// SetOutputText sets the OutputSpeech type to text and sets the value specified.
func (r *Response) SetOutputText(text string) {
	r.OutputSpeech = &OutputSpeech{Type: "PlainText", Text: text}
}

// SetOutputSSML sets the OutputSpeech type to ssml and sets the value specified.
func (r *Response) SetOutputSSML(ssml string) {
	r.OutputSpeech = &OutputSpeech{Type: "SSML", SSML: ssml}
}

// SetRepromptText created a Reprompt if needed and sets the OutputSpeech type to text and sets the value specified.
func (r *Response) SetRepromptText(text string) {
	if r.Reprompt == nil {
		r.Reprompt = &Reprompt{}
	}
	r.Reprompt.OutputSpeech = &OutputSpeech{Type: "PlainText", Text: text}
}

// SetRepromptSSML created a Reprompt if needed and sets the OutputSpeech type to ssml and sets the value specified.
func (r *Response) SetRepromptSSML(ssml string) {
	if r.Reprompt == nil {
		r.Reprompt = &Reprompt{}
	}
	r.Reprompt.OutputSpeech = &OutputSpeech{Type: "SSML", SSML: ssml}
}

// verifyApplicationId verifies that the ApplicationID sent in the request
// matches the one configured for this skill.
func (alexa *Alexa) verifyApplicationID(request *RequestEnvelope) error {
	appID := alexa.ApplicationID
	requestAppID := request.Session.Application.ApplicationID
	if appID == "" {
		return errors.New("Application ID was set to an empty string.")
	}
	if requestAppID == "" {
		return errors.New("Request Application ID was set to an empty string.")
	}
	if appID != requestAppID {
		return errors.New("Request Application ID does not match expected ApplicationId")
	}

	return nil
}

// verifyTimestamp compares the request timestamp to the current timestamp
// and returns an error if they are too far apart.
func (alexa *Alexa) verifyTimestamp(request *RequestEnvelope) error {

	timestamp, err := time.Parse(time.RFC3339, request.Request.Timestamp)
	if err != nil {
		return errors.New("Unable to parse request timestamp.  Err: " + err.Error())
	}
	now := time.Now()
	delta := now.Sub(timestamp)
	deltaSecsAbs := math.Abs(delta.Seconds())
	if deltaSecsAbs > float64(timestampTolerance) {
		return errors.New("Invalid Timestamp. The request timestap " + timestamp.String() + " was off the current time " + now.String() + " by more than " + strconv.FormatInt(int64(timestampTolerance), 10) + " seconds.")
	}

	return nil
}
