// Webhook contains utils for running a simple, cancelable webhook server
package webhooks

import (
	"context"
	"fmt"
	"github.com/solo-io/autopilot/pkg/utils"
	"net/http"
)

// a webhook Receiver receives webhook payloads.
// it is initialized by 
type Receiver interface {
	// Payloads returns a channel of payloads received on the webhook
	Payloads() <-chan *http.Request
}

// The HttpWebhookReceiver acts as an http.Handler
// that pushes received payloads to a channel
type HttpWebhookReceiver struct {
	ctx      context.Context
	path     string
	payloads chan *http.Request
}

var _ Receiver = &HttpWebhookReceiver{}
var _ http.Handler = &HttpWebhookReceiver{}

func NewHttpWebhookReceiver(ctx context.Context, path string) *HttpWebhookReceiver {
	payloads := make(chan *http.Request)
	return &HttpWebhookReceiver{ctx: ctx, path: path, payloads: payloads}
}

func (r *HttpWebhookReceiver) Payloads() <-chan *http.Request {
	return r.payloads
}

func (r *HttpWebhookReceiver) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	logger := utils.LoggerFromContext(r.ctx).WithValues("webhook", r.path)
	logger.V(0).Info("received payload", "remote", request.RemoteAddr)
	select {
	case r.payloads <- request:
	case <-r.ctx.Done():
		return
	}
}

// convenience function to create the  and listen on it in a goroutine. the started http server will be closed when the context is cancelled.
func MakeReceiverServer(ctx context.Context, port uint32, path string) (*http.Server, Receiver) {
	payloads := make(chan *http.Request)
	receiver := &HttpWebhookReceiver{ctx: ctx, path: path, payloads: payloads}
	mux := http.NewServeMux()
	mux.Handle(path, receiver)
	return &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,
	}, receiver
}
