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
