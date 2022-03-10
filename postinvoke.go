package postinvoke

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"os/signal"
	"syscall"
)

const Address = "127.0.0.1:1339"

type handler struct {
	inner  lambda.Handler
	client *http.Client
	stack  []func()
}

type contextKey string

const contextKeyWrapper = contextKey("contextKeyWrapper")

type Options struct {
	Client *http.Client
}

func WrapHandler(inner lambda.Handler, opts *Options) lambda.Handler {
	if opts == nil {
		opts = &Options{}
	}

	c := opts.Client
	if c == nil {
		c = http.DefaultClient
	}

	h := &handler{inner: inner, client: c}

	get, err := h.client.Get(fmt.Sprintf("http://%s/check", Address))
	if err != nil || get.StatusCode != 200 {
		panic("postinvoke: unable to connect to extension - did you forget to add the lambda layer?")
	}

	return h
}

func (h *handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	ctx = context.WithValue(ctx, contextKeyWrapper, h)
	defer func() {
		go h.after()
	}()

	return h.inner.Invoke(ctx, payload)
}

func (h *handler) after() {
	for i := len(h.stack) - 1; i >= 0; i-- {
		fn := h.stack[i]
		fn()
	}

	// empty the stack, keep the memory
	h.stack = h.stack[:0]

	_, err := h.client.Post(fmt.Sprintf("http://%s/done", Address), "", nil)
	if err != nil {
		panic(err)
	}
}

func Run(ctx context.Context, fn func()) {
	wrapper, ok := ctx.Value(contextKeyWrapper).(*handler)
	if !ok {
		panic("postinvoke: context unavailable - did you forget to wrap your lambda handler?")
	}

	wrapper.stack = append(wrapper.stack, fn)
}

func Shutdown(fn func()) {
	go func() {
		ctx := context.Background()
		ctx, _ = signal.NotifyContext(ctx, syscall.SIGTERM)
		<-ctx.Done()
		fn()
	}()
}
