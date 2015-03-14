package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	r "github.com/dancannon/gorethink"
	"github.com/lavab/api/models"
	"github.com/lavab/api/utils"
)

type createInput struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type createMsg struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
}

func create(w http.ResponseWriter, req *http.Request) {
	// Decode the body
	var msg createInput
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

	// Normalize the username
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
		writeJSON(w, errorMsg{
			Success: false,
			Message: "Username is taken",
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
		writeJSON(w, errorMsg{
			Success: false,
			Message: "Email is already used",
		})
		return
	}

	// Get account if invite.AccountID is set
	var account *models.Account
	if invite.AccountID != "" {
		cursor, err = r.Db(*rethinkAPIName).Table("accounts").Get(invite.AccountID).Run(session)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		err = cursor.One(&account)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		if account.Name != msg.Username || account.AltEmail != msg.Email {
			err = r.Db(*rethinkAPIName).Table("accounts").Get(invite.AccountID).Update(map[string]interface{}{
				"name":      msg.Username,
				"alt_email": msg.Email,
			}).Exec(session)
			if err != nil {
				writeJSON(w, errorMsg{
					Success: false,
					Message: err.Error(),
				})
				return
			}
		}
	} else {
		// Prepare a new account
		account = &models.Account{
			Resource: models.MakeResource("", msg.Username),
			AltEmail: msg.Email,
			Status:   "registered",
			Type:     "beta",
		}

		// Update the invite
		invite.AccountID = account.ID
		err = r.Db(*rethinkName).Table("invites").Get(invite.ID).Update(map[string]interface{}{
			"account_id": invite.AccountID,
		}).Exec(session)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Insert the account into db
		err = r.Db(*rethinkAPIName).Table("accounts").Insert(account).Exec(session)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}
	}

	// Generate a new invite token for the user
	token := &models.Token{
		Resource: models.MakeResource(account.ID, "Invitation token from invite-api"),
		Type:     "verify",
		Expiring: models.Expiring{
			ExpiryDate: time.Now().UTC().Add(time.Hour * 12),
		},
	}

	// Insert it into db
	err = r.Db(*rethinkAPIName).Table("tokens").Insert(token).Exec(session)
	if err != nil {
		writeJSON(w, errorMsg{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Return the token
	writeJSON(w, createMsg{
		Success: true,
		Code:    token.ID,
	})
}
