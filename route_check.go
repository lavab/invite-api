package main

import (
	"encoding/json"
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/lavab/api/models"
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

	resp := checkMsg{
		Success: true,
	}

	// If AccountID is set, then fetch the account from the database
	if invite.AccountID != "" {
		cursor, err = r.Db(*rethinkAPIName).Table("accounts").Get(invite.AccountID).Run(session)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		var account *models.Account
		err = cursor.One(&account)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// And put the result in the response
		resp.Email = account.AltEmail
		resp.Name = account.Name
	}

	// Respond using the invite
	writeJSON(w, resp)
}
