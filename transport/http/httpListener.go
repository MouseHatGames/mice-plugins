package http

import (
	"io/ioutil"
	"net/http"

	"github.com/MouseHatGames/mice/transport"
)

type httpListener struct {
}

func (l *httpListener) Accept(fn func(transport.Socket)) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		fn(&httpSocket{r, body})
	})

	srv := &http.Server{Handler: mux}

	return srv.ListenAndServe()
}
