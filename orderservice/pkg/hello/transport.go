package hello

import (
	"fmt"
	"net/http"
)

func GetHelloWorld(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "Hello World")
}
