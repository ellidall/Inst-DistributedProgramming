package kitty

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Kitty struct {
	Name string `json:"name"`
}

func GetKitty(w http.ResponseWriter, _ *http.Request) {
	cat := Kitty{Name: "Kitty"}
	body, err := json.Marshal(cat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err = w.Write(body); err != nil {
		log.WithField("err", err).Error("write response error")
	}
}
