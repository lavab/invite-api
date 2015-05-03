package main

import (
	"net/http"
)

func info(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("info route"))
}
