# HelloWorld Sample

This SDK was designed to be used as an AWS Lambda function via the [eawsy lambda shim](https://github.com/eawsy/aws-lambda-go-shim).  Please review their
documentation and install the necessary dependencies if you will be deploying
this sample to AWS Lambda.

This also assumes you have the [Amazon AWS CLI](https://aws.amazon.com/cli/) installed and configured.

First, create an Alexa Skill following the instructions described in the [Java HelloWorld Sample](https://github.com/amzn/alexa-skills-kit-java/tree/master/samples/src/main/java/helloworld)

Second, compile the sample using the included Makefile

```
make all
```

Then, create a new Lambda function using the AWS CLI:

```
aws lambda create-function \
  --role arn:aws:iam::AWS_ACCOUNT_NUMBER:role/lambda_basic_execution \
  --function-name HelloWorld \
  --zip-file fileb://package.zip \
  --runtime python2.7 \
  --handler handler.Handle
```

You can now test the HelloWorld skill via an Echo attached to your Amazon account or using the Amazon Alexa Console.

Once the lambda function is created, you can use the make file to build and
update your function.

```
make all push
```
