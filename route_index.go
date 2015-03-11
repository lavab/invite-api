package main

import (
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("lavab/invite-api 0.1.0"))
}
