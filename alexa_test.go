package alexa

import (
	"context"
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
	ctx := context.Background()
	_, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Expected ProcessRequest to succeed but got error", err)
	}

	alexa = &Alexa{ApplicationID: "amzn1.ask.skill.ABC123456", RequestHandler: &emptyRequestHandler{}}
	_, err = alexa.ProcessRequest(ctx, request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail due to invalid Application ID but no err was returned.")
	}

	alexa = &Alexa{ApplicationID: "", RequestHandler: &emptyRequestHandler{}}
	_, err = alexa.ProcessRequest(ctx, request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail due to an empty Application ID but no err was returned.")
	}

	alexa = &Alexa{ApplicationID: applicationID, RequestHandler: &emptyRequestHandler{}}
	request.Session.Application.ApplicationID = ""
	_, err = alexa.ProcessRequest(ctx, request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail due to an empty Request Application ID but no err was returned.")
	}
}

func TestAlexaTimestampValidation(t *testing.T) {
	request := createRecipieRequest()

	alexa := getAlexa()
	duration, _ := time.ParseDuration("-145s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	ctx := context.Background()
	_, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Expected ProcessRequest to succeed but got error", err)
	}

	duration, _ = time.ParseDuration("-151s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err = alexa.ProcessRequest(ctx, request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail to due to an invalid timetamp but no err was returned.")
	}

	duration, _ = time.ParseDuration("151s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err = alexa.ProcessRequest(ctx, request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail to due to an invalid timetamp but no err was returned.")
	}

	request.Request.Timestamp = "UNPARSEABLE"
	_, err = alexa.ProcessRequest(ctx, request)
	if err == nil {
		t.Error("Expected ProcessRequest to fail because the timestamp could not be parsed but no err was returned")
	}

	alexa.SetTimestampTolerance(0)
	duration, _ = time.ParseDuration("-1s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	_, err = alexa.ProcessRequest(ctx, request)
	alexa.SetTimestampTolerance(150)
	if err == nil {
		t.Error("Expected ProcessRequest to fail to due to an invalid timetamp but no err was returned.")
	}

	// Test Disabled Timestamp
	duration, _ = time.ParseDuration("151s")
	request.Request.Timestamp = time.Now().Add(duration).Format(time.RFC3339)
	alexa.IgnoreTimestamp = true
	_, err = alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Expected ProcessRequest to pass even with an invalid timestamp because validation is disabled.")
	}

}

func TestAlexaOnSessionStartedCalled(t *testing.T) {
	request := createRecipieRequest()

	handler := &emptyRequestHandler{}
	alexa := getAlexaWithHandler(handler)
	ctx := context.Background()
	_, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if handler.OnSessionStartedCalled {
		t.Error("On SessionStarted was called when session was not new.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	request.Session.New = true
	_, err = alexa.ProcessRequest(ctx, request)
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
	_, err = alexa.ProcessRequest(ctx, request)
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
	ctx := context.Background()
	_, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnLaunchCalled {
		t.Error("On Launch was not called.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	handler.OnLaunchThrowsErr = true
	_, err = alexa.ProcessRequest(ctx, request)
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
	ctx := context.Background()
	_, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnIntentCalled {
		t.Error("OnIntent was not called.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	handler.OnIntentThrowsErr = true
	_, err = alexa.ProcessRequest(ctx, request)
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

	ctx := context.Background()
	_, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnSessionEndedCalled {
		t.Error("OnSessionEnded was not called.")
	}

	handler = &emptyRequestHandler{}
	alexa = getAlexaWithHandler(handler)
	handler.OnSessionEndedThrowsErr = true
	_, err = alexa.ProcessRequest(ctx, request)
	if !handler.OnSessionEndedCalled {
		t.Error("OnSessionEnded was not called.")
	}
	if err == nil {
		t.Error("OnSessionEnded should have returned an error.")
	}
}

func TestAlexaSessionAttributesSet(t *testing.T) {
	request := createRecipieRequest()
	request.Request.Type = intentRequestName

	handler := &emptyRequestHandler{}
	handler.OnIntentSetsSessionAttr = true
	alexa := getAlexaWithHandler(handler)
	ctx := context.Background()
	resp, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if !handler.OnIntentCalled {
		t.Error("OnIntent was not called.")
	}
	if resp.SessionAttributes["myNewAttr"] != "Set123" {
		t.Error("Session Attribute myNewAttr should be Set123 in ResponseEnvelope but was", resp.SessionAttributes["myNewAttr"])
	}

}

func TestAlexaSimpleTextResponse(t *testing.T) {
	request := createRecipieRequest()

	alexa := getAlexaWithHandler(&simpleResponseHandler{})
	ctx := context.Background()
	responseEnv, err := alexa.ProcessRequest(ctx, request)
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
	ctx := context.Background()
	responseEnv, err := alexa.ProcessRequest(ctx, request)
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
	ctx := context.Background()
	responseEnv, err := alexa.ProcessRequest(ctx, request)
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
	responseEnv, err = alexa.ProcessRequest(ctx, request)
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
	responseEnv, err = alexa.ProcessRequest(ctx, request)
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
	ctx := context.Background()
	responseEnv, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if len(responseEnv.Response.Directives) != 1 {
		t.Fatalf("Response should contain 1 directive but contains %d", len(responseEnv.Response.Directives))
	}

	exp := `{"type":"AudioPlayer.Play","playBehavior":"REPLACE_ALL","audioItem":{"stream":{"token":"track2-long-audio","url":"https://my-audio-hosting-site.com/audio/sample-song-2.mp3","offsetInMilliseconds":100}}}`

	b, err := json.Marshal(responseEnv.Response.Directives[0])
	if err != nil {
		t.Fatalf("Error marshaling response. %s", err.Error())
	}
	if string(b) != exp {
		t.Errorf("Expected JSON of "+exp+" but was %s", string(b))
	}
}

func TestSimpleDialogDirective(t *testing.T) {
	request := createRecipieRequest()

	simpleDialogDirectiveResponseHandler := &simpleDialogDirectiveResponseHandler{Type: "Simple"}
	alexa := getAlexaWithHandler(simpleDialogDirectiveResponseHandler)
	ctx := context.Background()
	responseEnv, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if len(responseEnv.Response.Directives) != 1 {
		t.Fatalf("Response should contain 1 directive but contains %d", len(responseEnv.Response.Directives))
	}

	exp := `{"type":"Dialog.Delegate","updatedIntent":{"name":"PlanMyTrip","confirmationStatus":"NONE","slots":{"travelDate":{"name":"travelDate","confirmationStatus":"NONE","value":"2017-04-21"}}}}`

	b, err := json.Marshal(responseEnv.Response.Directives[0])
	if err != nil {
		t.Fatalf("Error marshaling response. %s", err.Error())
	}
	if string(b) != exp {
		t.Errorf("Expected JSON of "+exp+" but was %s", string(b))
	}
}

func TestNoIntentDialogDirective(t *testing.T) {
	request := createRecipieRequest()

	simpleDialogDirectiveResponseHandler := &simpleDialogDirectiveResponseHandler{Type: "NoIntent"}
	alexa := getAlexaWithHandler(simpleDialogDirectiveResponseHandler)
	ctx := context.Background()
	responseEnv, err := alexa.ProcessRequest(ctx, request)
	if err != nil {
		t.Error("Error processing request. " + err.Error())
	}
	if len(responseEnv.Response.Directives) != 1 {
		t.Fatalf("Response should contain 1 directive but contains %d", len(responseEnv.Response.Directives))
	}

	exp := `{"type":"Dialog.Delegate"}`

	b, err := json.Marshal(responseEnv.Response.Directives[0])
	if err != nil {
		t.Fatalf("Error marshaling response. %s", err.Error())
	}
	if string(b) != exp {
		t.Errorf("Expected JSON of "+exp+" but was %s", string(b))
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
	OnIntentSetsSessionAttr bool
	OnSessionEndedCalled    bool
	OnSessionEndedThrowsErr bool
}

func (h *emptyRequestHandler) OnSessionStarted(context.Context, *Request, *Session, *Response) error {
	h.OnSessionStartedCalled = true
	if h.OnSessionStartThrowsErr {
		return errors.New("Error in OnSessionStarted")
	}
	return nil
}

func (h *emptyRequestHandler) OnLaunch(context.Context, *Request, *Session, *Response) error {
	h.OnLaunchCalled = true
	if h.OnLaunchThrowsErr {
		return errors.New("Error in OnLaunch")
	}
	return nil
}

func (h *emptyRequestHandler) OnIntent(c context.Context, req *Request, s *Session, res *Response) error {
	h.OnIntentCalled = true
	if h.OnIntentSetsSessionAttr {
		s.Attributes.String["myNewAttr"] = "Set123"
	}
	if h.OnIntentThrowsErr {
		return errors.New("Error in OnIntent")
	}
	return nil
}

func (h *emptyRequestHandler) OnSessionEnded(context.Context, *Request, *Session, *Response) error {
	h.OnSessionEndedCalled = true
	if h.OnSessionEndedThrowsErr {
		return errors.New("Error in OnSessionEnded")
	}
	return nil
}

type simpleResponseHandler struct {
}

func (h *simpleResponseHandler) OnSessionStarted(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleResponseHandler) OnLaunch(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleResponseHandler) OnIntent(context context.Context, request *Request, session *Session, response *Response) error {

	response.SetOutputText("Response Text")
	response.SetRepromptText("Reprompt Text")

	return nil
}

func (h *simpleResponseHandler) OnSessionEnded(context.Context, *Request, *Session, *Response) error {
	return nil
}

type simpleSSMLResponseHandler struct {
}

func (h *simpleSSMLResponseHandler) OnSessionStarted(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleSSMLResponseHandler) OnLaunch(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleSSMLResponseHandler) OnIntent(context context.Context, request *Request, session *Session, response *Response) error {

	response.SetOutputSSML("<speak>This output speech uses SSML.</speak>")
	response.SetRepromptSSML("<speak>This Reprompt speech uses SSML.</speak>")

	return nil
}

func (h *simpleSSMLResponseHandler) OnSessionEnded(context.Context, *Request, *Session, *Response) error {
	return nil
}

type simpleCardResponseHandler struct {
	Type string
}

func (h *simpleCardResponseHandler) OnSessionStarted(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleCardResponseHandler) OnLaunch(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleCardResponseHandler) OnIntent(context context.Context, request *Request, session *Session, response *Response) error {

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

func (h *simpleCardResponseHandler) OnSessionEnded(context.Context, *Request, *Session, *Response) error {
	return nil
}

type simpleAudioPlayerResponseHandler struct {
	Type string
}

func (h *simpleAudioPlayerResponseHandler) OnSessionStarted(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleAudioPlayerResponseHandler) OnLaunch(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleAudioPlayerResponseHandler) OnIntent(context context.Context, request *Request, session *Session, response *Response) error {

	response.AddAudioPlayer("AudioPlayer.Play", "REPLACE_ALL", "track2-long-audio", "https://my-audio-hosting-site.com/audio/sample-song-2.mp3", 100)

	return nil
}

func (h *simpleAudioPlayerResponseHandler) OnSessionEnded(context.Context, *Request, *Session, *Response) error {
	return nil
}

type simpleDialogDirectiveResponseHandler struct {
	Type string
}

func (h *simpleDialogDirectiveResponseHandler) OnSessionStarted(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleDialogDirectiveResponseHandler) OnLaunch(context.Context, *Request, *Session, *Response) error {
	return nil
}

func (h *simpleDialogDirectiveResponseHandler) OnIntent(context context.Context, request *Request, session *Session, response *Response) error {

	switch h.Type {
	case "Simple":
		i := &Intent{
			Name:               "PlanMyTrip",
			ConfirmationStatus: "NONE",
			Slots: map[string]IntentSlot{
				"travelDate": IntentSlot{
					Name:               "travelDate",
					ConfirmationStatus: "NONE",
					Value:              "2017-04-21",
				},
			},
		}
		response.AddDialogDirective("Dialog.Delegate", "", "", i)
	case "NoIntent":
		response.AddDialogDirective("Dialog.Delegate", "", "", nil)
	}

	return nil
}

func (h *simpleDialogDirectiveResponseHandler) OnSessionEnded(context.Context, *Request, *Session, *Response) error {
	return nil
}
