package main

import (
	"time"
)

type Invite struct {
	ID        string    `gorethink:"id"`
	Source    string    `gorethink:"source"`
	AccountID string    `gorethink:"account_id,omitempty"` // maps to api.account:id
	CreatedBy string    `gorethink:"created_by,omitempty"` // maps to user:id
	CreatedAt time.Time `gorethink:"created_at"`
}
