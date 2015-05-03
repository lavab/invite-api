package main

import (
	"encoding/json"
	"net/http"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/dchest/uniuri"
	"github.com/lavab/api/models"
	"github.com/wunderlist/ttlcache"
)

type authInput struct {
	Username string
	Password string
}

type authMsg struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}

var tokens = ttlcache.NewCache(time.Hour * 12)

func auth(w http.ResponseWriter, req *http.Request) {
	var msg authInput
	err := json.NewDecoder(req.Body).Decode(&msg)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Fetch account from database
	cursor, err := r.Db(*rethinkAPIName).Table("accounts").GetAllByIndex("name", msg.Username).Run(session)
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

	// Verify password from input
	valid, nu, err := account.VerifyPassword(msg.Password)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if !valid {
		writeJSON(w, errorMsg{
			Success: false,
			Message: "Invalid password",
		})
		return
	}
	if nu {
		if err := r.Db(*rethinkAPIName).Table("accounts").Get(account.ID).Update(account).Exec(session); err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}
	}

	// Valid username, valid password - create a new token
	token := uniuri.NewLen(uniuri.UUIDLen)
	tokens.Set(token, account.ID)

	writeJSON(w, authMsg{
		Success: true,
		Token:   token,
	})
}
