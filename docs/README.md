# `postinvoke` for Go on AWS Lambda

AWS Lambda is a neat service. And it got neater with the launch of [external extensions][ext],
which allow you to have background processes that can continue running (e.g. for 
cleanup) after a  Lambda function has returned its response. But sometimes you 
want to do cleanup in the function process itself. Here's a library to achieve that.

## Usage

First, create a Lambda function. You need to include the supporting `postinvoke`
external extension in a layer. Here's an example that you can use (uncomment
as appropriate for your architecture):

```yaml
Transform: AWS::Serverless-2016-10-31

Resources:
  Example:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: ./example/bootstrap
      Runtime: provided.al2
      Handler: unused
      Timeout: 30
      MemorySize: 512
      Architectures: [arm64]
      Layers:
        - !Sub arn:aws:lambda:${AWS::Region}:514202201242:layer:postinvoke-arm64:1
#      Architectures: [x86_64]
#      Layers:
#        - !Sub arn:aws:lambda:${AWS::Region}:514202201242:layer:postinvoke-x86_64:1
```

Second, here's the code for an example Lambda function:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aidansteele/postinvoke"
	"github.com/aws/aws-lambda-go/lambda"
	"time"
)

func main() {
	postinvoke.Shutdown(func() {
		// this will be executed when the lambda environment shuts down.
		// you get 300ms to clean up - be quick!
		fmt.Println("bye y'all!")
	})

	h := lambda.NewHandler(handle)
	h = postinvoke.WrapHandler(h, nil) // don't forget this line
	lambda.StartHandler(h)
}

func handle(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	// this is where you do your lambda thing
	fmt.Println(string(input))

	postinvoke.Run(ctx, func() {
		// and here is where you can do your post-invoke cleanup. this code runs
		// after the lambda function has returned its response to the caller
		fmt.Println("first on stack")
	})

	postinvoke.Run(ctx, func() {
		// you can have multiple post-invoke methods
		// executed sequentially, in a defer-like stack
		fmt.Println("second on stack")
		time.Sleep(3 * time.Second)
		fmt.Println("second on complete")
	})

	return input, nil
}
```

This is what is logged by the above Lambda function. Note from the timestamps that
the environment is shut down about six minutes later.

![screenshot](/docs/screenshot.png)

[ext]: https://aws.amazon.com/blogs/compute/introducing-aws-lambda-extensions-in-preview/
