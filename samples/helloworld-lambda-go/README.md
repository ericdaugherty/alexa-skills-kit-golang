# HelloWorld Sample using aws-lambda-go

This sample is made for aws-lambda-go, lambda go runtime from sample helloworld that via shim. 

Support aws-lambda-go style kick handlers.
This is able to work with ask-cli.

This also assumes you have the [Amazon AWS CLI](https://aws.amazon.com/cli/) installed and configured. You should also have the "lambda_basic_execution"
role.

First, create an Alexa Skill following the instructions described in the [Java HelloWorld Sample](https://github.com/amzn/alexa-skills-kit-java/tree/master/samples/src/main/java/helloworld)

Second, retrieve dependencies

```
go get -u github.com/aws/aws-lambda-go/lambda
go get -u github.com/ericdaugherty/alexa-skills-kit-golang
```

Third, compile the sample using the included Makefile

```
make
```

Then, create a new Lambda function using the included Makefile:

```
make create
```

You can now test the HelloAlexa skill via an Echo attached to your Amazon account or using the Amazon Alexa Console.

Once the lambda function is created, you can use the Makefile to build and
update your function.

```
make
make update
```
