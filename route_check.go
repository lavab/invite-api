package main

import (
	"encoding/json"
	"net/http"

	r "github.com/dancannon/gorethink"
)

type checkInput struct {
	Token string `json:"token"`
}

type checkMsg struct {
	Success bool   `json:"success"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
}

func check(w http.ResponseWriter, req *http.Request) {
	// Decode the POST body
	var msg checkInput
	err := json.NewDecoder(req.Body).Decode(&msg)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Fetch the invite from database
	cursor, err := r.Db(*rethinkName).Table("invites").Get(msg.Token).Run(session)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	var invite *Invite
	err = cursor.One(&invite)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Respond using the invite
	writeJSON(w, checkMsg{
		Success: true,
		Name:    invite.Name,
		Email:   invite.Email,
	})
}
