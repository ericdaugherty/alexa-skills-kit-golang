# HelloWorld Sample

This SDK was designed to be used as an AWS Lambda function via the [eawsy lambda shim](https://github.com/eawsy/aws-lambda-go-shim).  Please review their
documentation and install the necessary dependencies if you will be deploying
this sample to AWS Lambda.

This also assumes you have the [Amazon AWS CLI](https://aws.amazon.com/cli/) installed and configured. You should also have the "lambda_basic_execution"
role.

First, create an Alexa Skill following the instructions described in the [Java HelloWorld Sample](https://github.com/amzn/alexa-skills-kit-java/tree/master/samples/src/main/java/helloworld)

Second, retrieve dependencies

```
docker pull eawsy/aws-lambda-go-shim:latest
go get -u -d github.com/eawsy/aws-lambda-go-core/...
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
