#HelloWorld

This sample provides an outline on how to develop and deploy an Alexa application in Go using the [AWS CLI](https://aws.amazon.com/cli/) and [ASK CLI](https://developer.amazon.com/docs/smapi/quick-start-alexa-skills-kit-command-line-interface.html) (Command Line Interface). Please install these tools before proceeding.

This Sample assumes a valid Go install and basic working knowledge in Go.

First, we will compile and deploy the Go HelloWorld code as an AWS Lambda Function.

Compile the HelloWorld Go code and upload it as an AWS Lambda Function.

Compile and package:
```
GOARCH=amd64 GOOS=linux go build -o helloworld
zip helloworld.zip helloworld
```

Create/upload AWS Lambda Function:

First, we need to determine the Role ARN to use. Ideally, we can use the lambda_basic_execution role. Execute the following command to get the ARN for the lambda_basic_execution role.

```
aws iam get-role --role-name lambda_basic_execution --query 'Role.Arn' --output text
```

Now, create the function, using the lambda_basic_execution ARN you found above in place of <ROLE_ARN>

```
aws lambda create-function \
  --function-name HelloAlexa \
  --zip-file fileb://helloworld.zip \
  --role <ROLE_ARN> \
  --runtime go1.x \
  --handler helloworld
```

This should return a JSON snippet. Make note of the FunctionArn.

We need to make this Lamdba callable from an Alexa Skill by adding a permission:
```
aws lambda add-permission \
  --function-name HelloAlexa \
  --statement-id "1234" \
  --action "lambda:InvokeFunction" \
  --principal "alexa-appkit.amazon.com"
```

Now we can create a new Alexa Skill using the provides skills.json definition. You will need to edit the skill.json file to reflect the correct Lamda ARN.

Edit the line below and replace 'TBD' with the correct value returned when you crated the Lambda function.
```
"uri": "arn:aws:lambda:TBD"
```

Now, create the skill:
```
ask api create-skill -f skill.json
```

Confirm the success of the creation by calling the get-skill-status as specified in the response:
```
ask api get-skill-status -s amzn1.ask.skill.<SKILL ID>
```

Now we need to define the interaction model used by this skill. A sample interaction model is provided and does not need to be modified for this example. You simply need to update the model as follows (note, replace the below skill id with the one returned in the previous command):

```
ask api update-model -s amzn1.ask.skill.<SKILL ID> -f interaction.json -l en-US
```

You can check the status again with ask api get-skill-status. It may take a moment for the interaction model to process.

Finally, enable the skill so you can test it:
```
ask api enable-skill -s amzn1.ask.skill.<SKILL ID>
```

## Test

You can now test the function by talking to your Alexa or using the [Alexa Console](https://developer.amazon.com/edw/home.html#/skills).

## Update

You can update the Go Lambda function by recompiling and zipping as above and calling:

```
aws lambda update-function-code \
  --function-name HelloAlexa \
  --zip-file fileb://helloworld.zip
```
