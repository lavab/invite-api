package main

import (
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
	r "github.com/dancannon/gorethink"
	"github.com/lavab/api/utils"
)

type freeInput struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type freeMsg struct {
	Success       bool `json:"success"`
	UsernameTaken bool `json:"username_taken,omitempty"`
	EmailUsed     bool `json:"email_used,omitempty"`
}

func free(w http.ResponseWriter, req *http.Request) {
	// Decode the POST body
	var msg freeInput
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

	// Normalize the username - make it lowercase and remove dots
	msg.Username = utils.RemoveDots(utils.NormalizeUsername(msg.Username))

	// Validate the email
	if !govalidator.IsEmail(msg.Email) {
		writeJSON(w, errorMsg{
			Success: false,
			Message: "Invalid email address",
		})
		return
	}

	// Check if username is taken
	cursor, err = r.Db(*rethinkAPIName).Table("accounts").
		GetAllByIndex("name", msg.Username).
		Filter(r.Row.Field("id").Ne(r.Expr(invite.AccountID))).
		Count().Run(session)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	var usernameCount int
	err = cursor.One(&usernameCount)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if usernameCount > 0 {
		writeJSON(w, freeMsg{
			Success:       false,
			UsernameTaken: true,
		})
		return
	}

	// Check if email is used
	cursor, err = r.Db(*rethinkAPIName).Table("accounts").
		GetAllByIndex("alt_email", msg.Email).
		Filter(r.Row.Field("id").Ne(r.Expr(invite.AccountID))).
		Count().Run(session)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	var emailCount int
	err = cursor.One(&emailCount)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if emailCount > 0 {
		writeJSON(w, freeMsg{
			Success:   false,
			EmailUsed: true,
		})
		return
	}

	// Return the result
	writeJSON(w, freeMsg{
		Success: true,
	})
}
