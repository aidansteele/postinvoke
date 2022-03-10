package main

import (
	"context"
	"github.com/aidansteele/postinvoke"
	"github.com/aidansteele/postinvoke/layer/extension"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	ctx := context.Background()

	c := extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	name := filepath.Base(os.Args[0]) // extension name has to match the filename

	_, err := c.Register(ctx, name)
	if err != nil {
		panic(err)
	}

	ch := make(chan struct{})

	http.HandleFunc("/done", func(w http.ResponseWriter, r *http.Request) {
		ch <- struct{}{}
	})

	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		// used by sdk to check if the extension is running
		w.WriteHeader(200)
	})

	go http.ListenAndServe(postinvoke.Address, nil)
	loop(ctx, c, ch)
}

func loop(ctx context.Context, c *extension.Client, ch chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := c.NextEvent(ctx)
			if err != nil {
				panic(err)
			}

			// wait for sdk to say its done
			<-ch
		}
	}
}
