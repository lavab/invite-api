package main

import (
	"encoding/json"
	"io"
	"log"
)

type errorMsg struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func writeJSON(w io.Writer, value interface{}) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Print(err)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
		return
	}
}
