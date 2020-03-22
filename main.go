package main

import (
	"flag"
	"github.com/eyedeekay/blind-turn/src/client"
	"github.com/eyedeekay/blind-turn/src/server"
)

func main() {
	client := flag.Bool("client", false, "Run as a client and not a server")

	users := flag.String("users", "", "List of username and password (e.g. \"user=pass,user=pass\")")
	realm := flag.String("realm", "", "Realm (defaults to base32 used by the service for servers, mandatory for clients)")

	host := flag.String("host", "", "TURN Server b32.")
	user := flag.String("user", "", "A pair of username and password (e.g. \"user=pass\")")
	ping := flag.Bool("ping", false, "Run ping test")

	flag.Parse()
	if *client {
		blindclient.Main(*host, *user, *realm, *ping)
	} else {
		blindserver.Main(*users, *realm)
	}
}
