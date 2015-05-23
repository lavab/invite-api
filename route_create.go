package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	r "github.com/dancannon/gorethink"
	//"github.com/dchest/uniuri"
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
	styledName := msg.Username
	msg.Username = utils.RemoveDots(utils.NormalizeUsername(msg.Username))

	var account *models.Account

	// If there's no account id, then simply check args
	if invite.AccountID == "" {
		if !govalidator.IsEmail(msg.Email) {
			writeJSON(w, errorMsg{
				Success: false,
				Message: "Invalid email address",
			})
			return
		}

		// Check if address is taken
		cursor, err = r.Db(*rethinkAPIName).Table("addresses").Get(msg.Username).Run(session)
		if err == nil || cursor != nil {
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

		// Prepare a new account
		account = &models.Account{
			Resource:   models.MakeResource("", msg.Username),
			AltEmail:   msg.Email,
			StyledName: styledName,
			Status:     "registered",
			Type:       "supporter",
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
	} else {
		cursor, err = r.Db(*rethinkAPIName).Table("accounts").Get(invite.AccountID).Run(session)
		if err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		defer cursor.Close()
		if err := cursor.One(&account); err != nil {
			writeJSON(w, errorMsg{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		if account.Name != "" && account.Name != msg.Username {
			writeJSON(w, errorMsg{
				Success: false,
				Message: "Invalid username",
			})
			return
		} else if account.Name == "" {
			// Check if address is taken
			cursor, err = r.Db(*rethinkAPIName).Table("addresses").Get(msg.Username).Run(session)
			if err == nil || cursor != nil {
				writeJSON(w, errorMsg{
					Success: false,
					Message: "Username is taken",
				})
				return
			}
		}

		if account.AltEmail != "" && account.AltEmail != msg.Email {
			writeJSON(w, errorMsg{
				Success: false,
				Message: "Invalid email",
			})
			return
		}

		if account.AltEmail == "" {
			if !govalidator.IsEmail(msg.Email) {
				writeJSON(w, errorMsg{
					Success: false,
					Message: "Invalid email address",
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
			defer cursor.Close()

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
		}

		if err := r.Db(*rethinkAPIName).Table("accounts").Get(invite.AccountID).Update(map[string]interface{}{
			"type": "supporter",
		}).Exec(session); err != nil {
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

	// Here be dragons. Thou art forewarned.
	/*go func() {
		// Watch the changes
		cursor, err := r.Db(*rethinkAPIName).Table("accounts").Get(account.ID).Changes().Run(session)
		if err != nil {
			log.Print("Error while watching changes of user " + account.Name + " - " + err.Error())
			return
		}
		defer cursor.Close()

		// Generate a timeout "flag"
		ts := uniuri.New()

		// Read them
		c := make(chan struct{})
		go func() {
			var change struct {
				NewValue map[string]interface{} `gorethink:"new_val"`
			}
			for cursor.Next(&change) {
				if status, ok := change.NewValue["status"]; ok {
					if x, ok := status.(string); ok && x == "setup" {
						c <- struct{}{}
						return
					}
				}

				if iat, ok := change.NewValue["_invite_api_timeout"]; ok {
					if x, ok := iat.(string); ok && x == ts {
						log.Print("Account setup watcher timeout for name " + account.Name)
						return
					}
				}
			}
		}()

		// Block the goroutine
		select {
		case <-c:
			if err := r.Db(*rethinkName).Table("invites").Get(invite.ID).Delete().Exec(session); err != nil {
				log.Print("Unable to delete an invite. " + invite.ID + " - " + account.ID)
				return
			}
			return
		case <-time.After(12 * time.Hour):
			if err := r.Db(*rethinkAPIName).Table("accounts").Get(account.ID).Update(map[string]interface{}{
				"_invite_api_timeout": ts,
			}).Exec(session); err != nil {
				log.Print("Failed to make a goroutine timeout. " + account.ID)
			}
			return
		}
	}()*/

	// jk fuck that
	if err := r.Db(*rethinkName).Table("invites").Get(invite.ID).Delete().Exec(session); err != nil {
		log.Print("Unable to delete an invite. " + invite.ID + " - " + account.ID)
		return
	}

	// Return the token
	writeJSON(w, createMsg{
		Success: true,
		Code:    token.ID,
	})
}
