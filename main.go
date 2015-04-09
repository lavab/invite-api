package main

import (
	"log"
	"strings"

	r "github.com/dancannon/gorethink"
	"github.com/lavab/flag"
	"github.com/lavab/kiri"
	"github.com/rs/cors"
	"github.com/zenazn/goji"
)

var (
	rethinkAddress = flag.String("rethinkdb_address", "127.0.0.1:28015", "Address of the RethinkDB server")
	rethinkName    = flag.String("rethinkdb_name", "invite", "Name of the invitation app's database")
	rethinkKey     = flag.String("rethinkdb_key", "", "Key of the RethinkDB connection")
	rethinkAPIName = flag.String("rethinkdb_api_name", "prod", "Name of the API's database")

	kiriAddresses = flag.String("kiri_addresses", "", "Addresses of the etcd servers to use")

	kiriDiscoveryStores    = flag.String("kiri_discovery_stores", "", "Stores list for service discovery. Syntax: kind,path;kind,path")
	kiriDiscoveryRethinkDB = flag.String("kiri_discovery_rethinkdb", "rethinkdb", "Name of the RethinkDB service in SD")

	kiriThisStores  = flag.String("kiri_this_stores", "", "Stores list for http backend registering. Syntax: kind,path;kind,path")
	kiriThisName    = flag.String("kiri_this_name", "invite-api", "Name of the HTTP service")
	kiriThisAddress = flag.String("kiri_this_address", "", "Address of this service")

	session *r.Session
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	// Parse kiri addresses
	ka := strings.Split(*kiriAddresses, ",")

	// Set up kiri agent for discovery
	kd := kiri.New(ka)

	// Set up a kiri agent for backend
	kb := kiri.New(ka)

	// Register this service in kb and add stores
	kb.Register(*kiriThisName, *kiriThisAddress, nil)
	for i, store := range strings.Split(*kiriThisStores, ";") {
		parts := strings.Split(store, ",")
		if len(parts) != 2 {
			log.Fatalf("Invalid parts count in kiri_this_stores#%d", i)
		}

		var kind kiri.Format
		switch parts[0] {
		case "default":
			kind = kiri.Default
		case "puro":
			kind = kiri.Puro
		default:
			log.Fatalf("Invalid kind of store in kiri_this_stores#%d", i)
		}
		kb.Store(kind, parts[1])
	}

	// Add stores to kd
	for i, store := range strings.Split(*kiriDiscoveryStores, ";") {
		parts := strings.Split(store, ",")
		if len(parts) != 2 {
			log.Fatalf("Invalid parts count in kiri_discovery_stores#%d", i)
		}

		var kind kiri.Format
		switch parts[0] {
		case "default":
			kind = kiri.Default
		case "puro":
			kind = kiri.Puro
		default:
			log.Fatalf("Invalid kind of store in kiri_discovery_stores#%d", i)
		}
		kd.Store(kind, parts[1])
	}

	// Connect to RethinkDB
	var session *r.Session
	err := kd.Discover(*kiriDiscoveryRethinkDB, nil, kiri.DiscoverFunc(func(service *kiri.Service) error {
		var err error
		session, err = r.Connect(r.ConnectOpts{
			Address: service.Address,
		})
		return err
	}))
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

	// Add a CORS middleware
	goji.Use(cors.New(cors.Options{
		AllowCredentials: true,
	}).Handler)

	// Add routes to goji
	goji.Get("/", index)
	goji.Post("/check", check)
	goji.Post("/create", create)

	// Start the server
	goji.Serve()
}
