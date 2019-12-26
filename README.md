# ericdaugherty/alexa-skills-kit-golang

alexa-skills-kit-golang is a lightweight port of the Amazon [alexa-skills-kit-java](https://github.com/amzn/alexa-skills-kit-java)
SDK and Samples.

[![License](https://img.shields.io/badge/License-Apache%202.0-lightgrey.svg)](https://opensource.org/licenses/Apache-2.0)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/ericdaugherty/alexa-skills-kit-golang)

## Usage

This explanation assumes familiarity with with AWS Documentation.  Please
review [Developing an Alexa Skill as a Lambda Function](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/developing-an-alexa-skill-as-a-lambda-function) before proceeding. This SDK addresses some of the steps documented here for you, but you should be familiar with the entire process.

The samples directory provides example usage.

The Alexa struct is the initial interface point with the SDK.  Alexa must be
 initialized first.  The struct is defined as:

```Go
type Alexa struct {
    ApplicationID       string
    RequestHandler      RequestHandler
    IgnoreApplicationID bool
    IgnoreTimestamp     bool
}
```

The ApplicationID must match the ApplicationID defined in the Alexa Skills

The RequestHandler is an interface that must be implemented, and is called to handle requests.

IgnoreApplicationID and IgnoreTimestamp should be used during debugging to test with hard-coded requests.

Requests from Alexa should be passed into the Alexa.ProcessRequest method.

```Go
func (alexa *Alexa) ProcessRequest(context context.Context, requestEnv *RequestEnvelope) (*ResponseEnvelope, error)
```

This method takes the incoming request and validates it, and the calls the
appropriate callback methods on the RequestHandler interface implementation.

The ResponseEnvelope is returned and can be converted to JSON to be passed
back to the Alexa skill.

RequestHandler interface is defined as:
```Go
type RequestHandler interface {
	OnSessionStarted(context.Context, *Request, *Session, *Context, *Response) error
	OnLaunch(context.Context, *Request, *Session, *Context, *Response) error
	OnIntent(context.Context, *Request, *Session, *Context, *Response) error
	OnSessionEnded(context.Context, *Request, *Session, *Context, *Response) error
}
```

For a summary of these methods, please see the [Handling Requests Sent By Alexa](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/handling-requests-sent-by-alexa) documentation.

You can directly manipulate the Response struct, but it is not initialized by default and use of the connivence methods is recommended.

These methods include:
```Go
func (r *Response) SetSimpleCard(title string, content string)
func (r *Response) SetStandardCard(title string, text string, smallImageURL string, largeImageURL string)
func (r *Response) SetLinkAccountCard()
func (r *Response) SetOutputText(text string)
func (r *Response) SetOutputSSML(ssml string)
func (r *Response) SetRepromptText(text string)
func (r *Response) SetRepromptSSML(ssml string)
```

And more.  These methods handle initializing any required struts within the Response struct as well as setting all required fields.

## samples

[HelloWorld](https://github.com/ericdaugherty/alexa-skills-kit-golang/tree/master/samples/helloworld)

## Limitations

This version does not support use as a standalone web server as it does not implement
any of the HTTPS validation.  It was developed to be used as an AWS Lambda function
using AWS Labda Go support.
