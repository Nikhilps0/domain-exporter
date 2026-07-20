package server

import (
	"fmt"
	"net/http"
)

func DefaultPath(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to Domain Exporter")
}
