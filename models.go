package main

import (
	"time"
)

type Invite struct {
	ID        string    `gorethink:"id"`
	Name      string    `gorethink:"name,omitempty"`
	Email     string    `gorethink:"email"`
	Source    string    `gorethink:"source"`
	CreatedBy string    `gorethink:"created_by"` // maps to user:id
	CreatedAt time.Time `gorethink:"created_at"`
}
