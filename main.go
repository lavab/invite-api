package main

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/lavab/goji"
	"github.com/namsral/flag"
	"github.com/rs/cors"
)

var (
	rethinkAddress = flag.String("rethinkdb_address", "127.0.0.1:28015", "Address of the RethinkDB server")
	rethinkName    = flag.String("rethinkdb_name", "invite", "Name of the invitation app's database")
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
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create the database and tables
	r.DbCreate(*rethinkName).Exec(session)
	r.Db(*rethinkName).TableCreate("invites").Exec(session)
	r.Db(*rethinkName).Table("invites").IndexCreate("email").Exec(session)
	r.Db(*rethinkName).Table("invites").IndexCreate("name").Exec(session)

	// Add a CORS middleware
	goji.Use(cors.New(cors.Options{
		AllowCredentials: true,
	}).Handler)

	// Add routes to goji
	goji.Get("/", index)
	goji.Post("/check", check)
	goji.Post("/free", free)
	goji.Post("/create", create)

	// Start the server
	goji.Serve()
}
