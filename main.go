package main

import (
	"flag"
	"github.com/eyedeekay/blind-turn/src/server"
)

func main() {
	users := flag.String("users", "", "List of username and password (e.g. \"user=pass,user=pass\")")
	realm := flag.String("realm", "", "Realm (defaults to base32 user by the service)")
	flag.Parse()
	blindserver.Main(*users, *realm)
}
