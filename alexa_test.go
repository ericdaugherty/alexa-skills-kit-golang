package alexa

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

const applicationID = "amzn1.ask.skill.ABC123"

const recipeIntentString = `{
  "session": {
    "new": false,
    "sessionId": "amzn1.echo-api.session.[unique-value-here]",
    "attributes": {},
    "user": {
      "userId": "amzn1.ask.account.[unique-value-here]"
    },
    "application": {
      "applicationId": "amzn1.ask.skill.ABC123"
    }
  },
  "version": "1.0",
  "request": {
    "locale": "en-US",
    "timestamp": "2016-10-27T21:06:28Z",
    "type": "IntentRequest",
    "requestId": "amzn1.echo-api.request.xyz789",
    "intent": {
      "slots": {
        "Item": {
          "name": "Item",
          "value": "snowball"
        }
      },
      "name": "RecipeIntent"
    }
  },
  "context": {
    "AudioPlayer": {
      "playerActivity": "IDLE"
    },
    "System": {
      "device": {
        "supportedInterfaces": {
          "AudioPlayer": {}
        }
      },
      "application": {
        "applicationId": "amzn1.ask.skill.[unique-value-here]"
      },
      "user": {
        "userId": "amzn1.ask.account.[unique-value-here]"
      }
    }
  }
}`

// TestAlexaJSON Verifies that the Alexa Struct parses an Alexa JSON String correctly.
func TestAlexaJSON(t *testing.T) {
	request := createRecipieRequest()

	if request.Version != "1.0" {
		t.Error("Expected Request Version to be 1.0 but was", request.Version)
	}
	if request.Session.New {
		t.Error("Expected Session.new to be false but was true.")
	}
	if request.Session.User.UserID != "amzn1.ask.account.[unique-value-here]" {
		t.Error("Expected Session.User.UserId to be amzn1.ask.account.[unique-value-here] but was", request.Session.User.UserID)
	}
	if request.Request.RequestID != "amzn1.echo-api.request.xyz789" {
		t.Error("Expected request.Request.RequestID to be amzn1.echo-api.request.xyz789 but was", request.Request.RequestID)
	}
}

func TestAlexaAppIDValidation(t *testing.T) {
	request := createRecipieRequest()

	alexa := getAlexa()
	_, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Expected ProcessRequest to succeed but got error", err)
	}

	alexa = &Alexa{ApplicationID: "amzn1.ask.skill.ABC123456", RequestHandler: &emptyRequestHandler{}}
	_, err = alexa.ProcessRequest(request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail due to invalid Application ID but no err was returned.")
	}

	alexa = &Alexa{ApplicationID: "", RequestHandler: &emptyRequestHandler{}}
	_, err = alexa.ProcessRequest(request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail due to an empty Application ID but no err was returned.")
	}

	alexa = &Alexa{ApplicationID: applicationID, RequestHandler: &emptyRequestHandler{}}
	request.Session.Application.ApplicationID = ""
	_, err = alexa.ProcessRequest(request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail due to an empty Request Application ID but no err was returned.")
	}
}

func TestAlexaTimestampValidation(t *testing.T) {
	request := createRecipieRequest()

	alexa := getAlexa()
	duration, _ := time.ParseDuration("-145s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Expected ProcessRequest to succeed but got error", err)
	}

	duration, _ = time.ParseDuration("-151s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err = alexa.ProcessRequest(request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail to due to an invalid timetamp but no err was returned.")
	}

	duration, _ = time.ParseDuration("151s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err = alexa.ProcessRequest(request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail to due to an invalid timetamp but no err was returned.")
	}

	request.Request.Timestamp = "UNPARSEABLE"
	_, err = alexa.ProcessRequest(request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail because the timestamp could not be parsed but no err was returned")
	}

	alexa.SetTimestampTolerance(0)
	duration, _ = time.ParseDuration("-1s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err = alexa.ProcessRequest(request)
	alexa.SetTimestampTolerance(150)
	if err == nil {
		t.Error("Expected ProcessRequest to fail to due to an invalid timetamp but no err was returned.")
	}

	// Test Disabled Timestamp
	duration, _ = time.ParseDuration("151s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	alexa.IgnoreTimestamp = true
	_, err = alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Expected ProcessRequest to pass even with an invalid timestamp because validation is disabled.")
	}

}

func TestAlexaOnSessionStartedCalled(t *testing.T) {
	request := createRecipieRequest()

	handler := &emptyRequestHandler{}
	alexa := getAlexaWithHandler(handler)
	_, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if handler.OnSessionStartedCalled {
		t.Error("On SessionStarted was called when session was not new.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	request.Session.New = true
	_, err = alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnSessionStartedCalled {
		t.Error("On SessionStarted was not called for a new session.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	request.Session.New = true
	handler.OnSessionStartThrowsErr = true
	_, err = alexa.ProcessRequest(request)
	if !handler.OnSessionStartedCalled {
		t.Error("On SessionStarted was not called for a new session.")
	}
	if err == nil {
		t.Error("OnSessionStart should have returned an error.")
	}
}

func TestAlexaOnLaunchCalled(t *testing.T) {
	request := createRecipieRequest()
	request.Request.Type = launchRequestName

	handler := &emptyRequestHandler{}
	alexa := getAlexaWithHandler(handler)
	_, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnLaunchCalled {
		t.Error("On Launch was not called.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	handler.OnLaunchThrowsErr = true
	_, err = alexa.ProcessRequest(request)
	if !handler.OnLaunchCalled {
		t.Error("OnLaunch was not called.")
	}
	if err == nil {
		t.Error("OnLaunch should have returned an error.")
	}
}

func TestAlexaOnIntentCalled(t *testing.T) {
	request := createRecipieRequest()
	request.Request.Type = intentRequestName

	handler := &emptyRequestHandler{}
	alexa := getAlexaWithHandler(handler)
	_, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnIntentCalled {
		t.Error("OnIntent was not called.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	handler.OnIntentThrowsErr = true
	_, err = alexa.ProcessRequest(request)
	if !handler.OnIntentCalled {
		t.Error("OnIntent was not called.")
	}
	if err == nil {
		t.Error("OnIntent should have returned an error.")
	}
}

func TestAlexaOnSessionEndedCalled(t *testing.T) {
	request := createRecipieRequest()
	request.Request.Type = sessionEndedRequestName

	handler := &emptyRequestHandler{}
	alexa := getAlexaWithHandler(handler)
	_, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnSessionEndedCalled {
		t.Error("OnSessionEnded was not called.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	handler.OnSessionEndedThrowsErr = true
	_, err = alexa.ProcessRequest(request)
	if !handler.OnSessionEndedCalled {
		t.Error("OnSessionEnded was not called.")
	}
	if err == nil {
		t.Error("OnSessionEnded should have returned an error.")
	}
}

func TestAlexaSimpleTextResponse(t *testing.T) {
	request := createRecipieRequest()

	alexa := getAlexaWithHandler(&simpleResponseHandler{})
	responseEnv, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}

	if responseEnv.Response.OutputSpeech.Text != "Response Text" {
		t.Errorf("Response Text should have been %s but was %s", "Response Text", responseEnv.Response.OutputSpeech.Text)
	}
	if responseEnv.Response.OutputSpeech.Type != "PlainText" {
		t.Errorf("Response Type should have been %s but was %s", "PlainText", responseEnv.Response.OutputSpeech.Type)
	}

	if responseEnv.Response.Reprompt.OutputSpeech.Text != "Reprompt Text" {
		t.Errorf("Response Text should have been %s but was %s", "Reprompt Text", responseEnv.Response.OutputSpeech.Text)
	}
	if responseEnv.Response.Reprompt.OutputSpeech.Type != "PlainText" {
		t.Errorf("Response Type should have been %s but was %s", "PlainText", responseEnv.Response.OutputSpeech.Type)
	}
}

func TestSimpleSSMLResponse(t *testing.T) {
	request := createRecipieRequest()

	alexa := getAlexaWithHandler(&simpleSSMLResponseHandler{})
	responseEnv, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}

	if responseEnv.Response.OutputSpeech.SSML != "<speak>This output speech uses SSML.</speak>" {
		t.Errorf("Response Text should have been %s but was %s", "<speak>This output speech uses SSML.</speak>", responseEnv.Response.OutputSpeech.SSML)
	}
	if responseEnv.Response.OutputSpeech.Type != "SSML" {
		t.Errorf("Response Type should have been %s but was %s", "SSML", responseEnv.Response.OutputSpeech.Type)
	}

	if responseEnv.Response.Reprompt.OutputSpeech.SSML != "<speak>This Reprompt speech uses SSML.</speak>" {
		t.Errorf("Response Text should have been %s but was %s", "<speak>This Reprompt speech uses SSML.</speak>", responseEnv.Response.OutputSpeech.SSML)
	}
	if responseEnv.Response.Reprompt.OutputSpeech.Type != "SSML" {
		t.Errorf("Response Type should have been %s but was %s", "SSML", responseEnv.Response.OutputSpeech.Type)
	}
}

func TestCards(t *testing.T) {
	request := createRecipieRequest()

	cardHandler := &simpleCardResponseHandler{Type: "Simple"}
	alexa := getAlexaWithHandler(cardHandler)
	responseEnv, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if responseEnv.Response.Card.Type != "Simple" {
		t.Errorf("Card Type should be Simple but was %s", responseEnv.Response.Card.Type)
	}
	if responseEnv.Response.Card.Content != "Simple Content" {
		t.Errorf("Card Content should be 'Simple Content' but was %s", responseEnv.Response.Card.Content)
	}

	cardHandler.Type = "Standard"
	responseEnv, err = alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if responseEnv.Response.Card.Type != "Standard" {
		t.Errorf("Card Type should be Standard but was %s", responseEnv.Response.Card.Type)
	}
	if responseEnv.Response.Card.Text != "Standard Body Text" {
		t.Errorf("Card Content should be 'Standard Body Text' but was %s", responseEnv.Response.Card.Text)
	}

	cardHandler.Type = "LinkAccount"
	responseEnv, err = alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if responseEnv.Response.Card.Type != "LinkAccount" {
		t.Errorf("Card Type should be LinkAccount but was %s", responseEnv.Response.Card.Type)
	}

}

func TestAudioPlayer(t *testing.T) {
	request := createRecipieRequest()

	audioPlayerHandler := &simpleAudioPlayerResponseHandler{Type: "Simple"}
	alexa := getAlexaWithHandler(audioPlayerHandler)
	responseEnv, err := alexa.ProcessRequest(request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if len(responseEnv.Response.Directives) != 1 {
		t.Fatalf("Response should contain 1 directive but contains %d", len(responseEnv.Response.Directives))
	}
	if responseEnv.Response.Directives[0].Type != "AudioPlayer.Play" {
		t.Errorf("Type should be AudioPlayer.Play but was %s", responseEnv.Response.Directives[0].Type)
	}
	if responseEnv.Response.Directives[0].PlayBehavior != "REPLACE_ALL" {
		t.Errorf("Type should be REPLACE_ALL but was %s", responseEnv.Response.Directives[0].PlayBehavior)
	}
	if responseEnv.Response.Directives[0].AudioItem.Stream.Token != "track2-long-audio" {
		t.Errorf("Type should be track2-long-audio but was %s", responseEnv.Response.Directives[0].PlayBehavior)
	}
	if responseEnv.Response.Directives[0].AudioItem.Stream.URL != "https://my-audio-hosting-site.com/audio/sample-song-2.mp3" {
		t.Errorf("Type should be https://my-audio-hosting-site.com/audio/sample-song-2.mp3 but was %s", responseEnv.Response.Directives[0].AudioItem.Stream.URL)
	}
	if responseEnv.Response.Directives[0].AudioItem.Stream.OffsetInMilliseconds != 100 {
		t.Errorf("Type should be 100 but was %d", responseEnv.Response.Directives[0].AudioItem.Stream.OffsetInMilliseconds)
	}
}

func getAlexa() *Alexa {
	return &Alexa{ApplicationID: applicationID, RequestHandler: &emptyRequestHandler{}}
}

func getAlexaWithHandler(handler RequestHandler) *Alexa {
	return &Alexa{ApplicationID: applicationID, RequestHandler: handler}
}

func createRecipieRequest() *RequestEnvelope {
	var request RequestEnvelope
	var jsonBlob = []byte(recipeIntentString)
	json.Unmarshal(jsonBlob, &request)
	request.Request.Timestamp = time.Now().Format(time.RFC3339)
	return &request
}

type emptyRequestHandler struct {
	OnSessionStartedCalled  bool
	OnSessionStartThrowsErr bool
	OnLaunchCalled          bool
	OnLaunchThrowsErr       bool
	OnIntentCalled          bool
	OnIntentThrowsErr       bool
	OnSessionEndedCalled    bool
	OnSessionEndedThrowsErr bool
}

func (h *emptyRequestHandler) OnSessionStarted(*Request, *Session, *Response) error {
	h.OnSessionStartedCalled = true
	if h.OnSessionStartThrowsErr {
		return errors.New("Error in OnSessionStarted")
	}
	return nil
}

func (h *emptyRequestHandler) OnLaunch(*Request, *Session, *Response) error {
	h.OnLaunchCalled = true
	if h.OnLaunchThrowsErr {
		return errors.New("Error in OnLaunch")
	}
	return nil
}

func (h *emptyRequestHandler) OnIntent(*Request, *Session, *Response) error {
	h.OnIntentCalled = true
	if h.OnIntentThrowsErr {
		return errors.New("Error in OnIntent")
	}
	return nil
}

func (h *emptyRequestHandler) OnSessionEnded(*Request, *Session, *Response) error {
	h.OnSessionEndedCalled = true
	if h.OnSessionEndedThrowsErr {
		return errors.New("Error in OnSessionEnded")
	}
	return nil
}

type simpleResponseHandler struct {
}

func (h *simpleResponseHandler) OnSessionStarted(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleResponseHandler) OnLaunch(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleResponseHandler) OnIntent(request *Request, session *Session, response *Response) error {

	response.SetOutputText("Response Text")
	response.SetRepromptText("Reprompt Text")

	return nil
}

func (h *simpleResponseHandler) OnSessionEnded(*Request, *Session, *Response) error {
	return nil
}

type simpleSSMLResponseHandler struct {
}

func (h *simpleSSMLResponseHandler) OnSessionStarted(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleSSMLResponseHandler) OnLaunch(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleSSMLResponseHandler) OnIntent(request *Request, session *Session, response *Response) error {

	response.SetOutputSSML("<speak>This output speech uses SSML.</speak>")
	response.SetRepromptSSML("<speak>This Reprompt speech uses SSML.</speak>")

	return nil
}

func (h *simpleSSMLResponseHandler) OnSessionEnded(*Request, *Session, *Response) error {
	return nil
}

type simpleCardResponseHandler struct {
	Type string
}

func (h *simpleCardResponseHandler) OnSessionStarted(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleCardResponseHandler) OnLaunch(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleCardResponseHandler) OnIntent(request *Request, session *Session, response *Response) error {

	switch h.Type {
	case "Simple":
		response.SetSimpleCard("Simple Title", "Simple Content")
	case "Standard":
		response.SetStandardCard("Standard Title", "Standard Body Text", "http://small.url", "http://large.url")
	case "LinkAccount":
		response.SetLinkAccountCard()
	}
	response.SetOutputSSML("<speak>This output speech uses SSML.</speak>")
	response.SetRepromptSSML("<speak>This Reprompt speech uses SSML.</speak>")

	return nil
}

func (h *simpleCardResponseHandler) OnSessionEnded(*Request, *Session, *Response) error {
	return nil
}

type simpleAudioPlayerResponseHandler struct {
	Type string
}

func (h *simpleAudioPlayerResponseHandler) OnSessionStarted(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleAudioPlayerResponseHandler) OnLaunch(*Request, *Session, *Response) error {
	return nil
}

func (h *simpleAudioPlayerResponseHandler) OnIntent(request *Request, session *Session, response *Response) error {

	response.AddAudioPlayer("AudioPlayer.Play", "REPLACE_ALL", "track2-long-audio", "https://my-audio-hosting-site.com/audio/sample-song-2.mp3", 100)

	return nil
}

func (h *simpleAudioPlayerResponseHandler) OnSessionEnded(*Request, *Session, *Response) error {
	return nil
}
