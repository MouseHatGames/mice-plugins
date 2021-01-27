package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/MouseHatGames/mice/transport"
)

const headerPrefix = "X-Mice-%s"

type httpTransport struct {
	log logger.Logger
}

func Transport() options.Option {
	return func(o *options.Options) {
		o.Transport = &httpTransport{
			log: o.Logger.GetLogger("http"),
		}
	}
}

func (t *httpTransport) Listen(ctx context.Context, addr string) (transport.Listener, error) {
	t.log.Infof("listening on %s", addr)

	return &httpListener{addr, t.log}, nil
}

func (t *httpTransport) Dial(ctx context.Context, addr string) (transport.Socket, error) {
	t.log.Infof("dialing %s", addr)

	return &httpOutgoingSocket{
		address: addr,
		resp:    make(chan *http.Response, 1),
	}, nil
}

type httpListener struct {
	addr string
	log  logger.Logger
}

func (l *httpListener) Close() error {
	return nil
}

func (l *httpListener) Accept(ctx context.Context, fn func(transport.Socket)) error {
	handler := http.NewServeMux()
	handler.HandleFunc("/request", func(rw http.ResponseWriter, r *http.Request) {
		l.log.Debugf("got request from %s", r.RemoteAddr)

		fn(&httpIncomingSocket{
			rw:  rw,
			r:   r,
			log: l.log,
		})
	})

	l.log.Debugf("accepting connections")
	http.ListenAndServe(l.addr, handler)
	return nil
}

func getMiceHeaders(h http.Header) (mh map[string]string) {
	mh = make(map[string]string)

	for k, v := range h {
		if strings.HasPrefix(k, headerPrefix) {
			name := strings.TrimPrefix(k, headerPrefix)
			mh[name] = v[0]
		}
	}

	return
}
