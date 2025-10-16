package orderservice

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type OrderResponse struct {
	ID     string `json:"id"`
	Some   string `json:"some,omitempty"`
	Status string `json:"status"`
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["ID"]
	some := r.URL.Query().Get("some")

	log.WithFields(log.Fields{
		"id":   id,
		"some": some,
	}).Info("got a new request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(OrderResponse{
		Status: "ok",
		ID:     id,
		Some:   some,
	})
	if err != nil {
		return
	}
}
