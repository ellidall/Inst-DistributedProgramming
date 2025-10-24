package server

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"orderservise/pkg/hello"
	"orderservise/pkg/orderservice"
)

func Router() http.Handler {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()
	s.HandleFunc("/hello-world", hello.GetHelloWorld).Methods(http.MethodGet)
	s.HandleFunc("/cat", hello.GetKitty).Methods(http.MethodGet)
	s.HandleFunc("/order/{ID}", orderservice.GetOrder).Methods(http.MethodGet)
	return logMiddleware(r)
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"method":     r.Method,
			"url":        r.URL,
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.UserAgent(),
		}).Info("got a new request")
		h.ServeHTTP(w, r)
	})
}
