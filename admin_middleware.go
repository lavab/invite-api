package main

import (
	"net/http"
	"strings"

	"github.com/lavab/goji/web"
)

func middleware(c *web.C, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the Authorization header
		header := r.Header.Get("Authorization")
		if header == "" {
			w.WriteHeader(403)
			writeJSON(w, errorMsg{
				Success: false,
				Message: "Missing auth token",
			})
			return
		}

		// Split it into two parts
		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			w.WriteHeader(400)
			writeJSON(w, errorMsg{
				Success: false,
				Message: "Invalid authorization header",
			})
			return
		}

		// Get account ID
		account, ok := tokens.Get(headerParts[1])
		if !ok {
			w.WriteHeader(403)
			writeJSON(w, errorMsg{
				Success: false,
				Message: "Invalid token",
			})
			return
		}

		// Insert into env
		c.Env["account"] = account
		h.ServeHTTP(w, r)
	})
}
