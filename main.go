package main

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/lavab/flag"
	"github.com/zenazn/goji"
)

var (
	rethinkAddress = flag.String("rethinkdb_address", "127.0.0.1:28015", "Address of the RethinkDB server")
	rethinkName    = flag.String("rethinkdb_name", "invite", "Name of the invitation app's database")
	rethinkKey     = flag.String("rethinkdb_key", "", "Key of the RethinkDB connection")
	rethinkAPIName = flag.String("rethinkdb_api_name", "prod", "Name of the API's database")

	session *r.Session
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	// Connect to RethinkDB
	var err error
	session, err = r.Connect(r.ConnectOpts{
		Address: *rethinkAddress,
		AuthKey: *rethinkKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create the database and tables
	r.DbCreate(*rethinkName).Exec(session)
	r.Db(*rethinkName).TableCreate("invites").Exec(session)
	r.Db(*rethinkName).Table("invites").IndexCreate("email").Exec(session)
	r.Db(*rethinkName).Table("invites").IndexCreate("name").Exec(session)
	r.Db(*rethinkName).TableCreate("users").Exec(session)
	r.Db(*rethinkName).Table("users").IndexCreate("name").Exec(session)

	// Add routes to goji
	goji.Get("/", index)
	goji.Post("/check", check)
	goji.Post("/create", create)

	// Start the server
	goji.Serve()
}
