package blindserver

import (
	//"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/eyedeekay/firefox-static/sammy"
	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	"github.com/pion/turn/v2"
)

type I2PRelayAddressGenerator struct {
}

// Validate confirms that the RelayAddressGenerator is properly initialized
func (i *I2PRelayAddressGenerator) Validate() error {
	return nil
}

// Allocate a PacketConn (UDP) RelayAddress
func (i *I2PRelayAddressGenerator) AllocatePacketConn(network string, requestedPort int) (net.PacketConn, net.Addr, error) {
	sam, err := sam3.NewSAM("127.0.0.1:7657")
	if err != nil {
		return nil, nil, err
	}
	keys, err := sam.NewKeys()
	if err != nil {
		return nil, nil, err
	}
	stream, err := sam.NewDatagramSession(keys.Addr().Base32()[0:9], keys, sam3.Options_Small, 0)
	if err != nil {
		return nil, nil, err
	}
	return stream, keys.Addr(), fmt.Errorf("UDP is not yet supported")
}

// Allocate a Conn (TCP) RelayAddress
func (i *I2PRelayAddressGenerator) AllocateConn(network string, requestedPort int) (net.Conn, net.Addr, error) {
	//tcpListener, err := sammy.Sammy()
	sam, err := sam3.NewSAM("127.0.0.1:7657")
	if err != nil {
		return nil, nil, err
	}
	keys, err := sam.NewKeys()
	if err != nil {
		return nil, nil, err
	}
	stream, err := sam.NewStreamSession(keys.Addr().Base32()[0:9], keys, sam3.Options_Small)
	if err != nil {
		return nil, nil, err
	}
	//fmt.Println("Client: Connecting to " + server.Base32())
	tcpConn, err := stream.DialI2P("")
	return tcpConn, keys.Addr(), nil
}

func Main(users, realm string) {
	tcpListener, err := sammy.Sammy()
	if err != nil {
		log.Panicf("Failed to create TURN server listener: %s", err)
	}
	if realm == "" {
		realm = tcpListener.Addr().(i2pkeys.I2PAddr).Base32()
	}

	if len(users) == 0 {
		log.Fatalf("'users' is required")
	}

	// Create a TCP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any TCP listeners, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic

	// Cache -users flag for easy lookup later
	// If passwords are stored they should be saved to your DB hashed using turn.GenerateAuthKey
	usersMap := map[string][]byte{}
	for _, kv := range regexp.MustCompile(`(\w+)=(\w+)`).FindAllStringSubmatch(users, -1) {
		usersMap[kv[1]] = turn.GenerateAuthKey(kv[1], realm, kv[2])
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: realm,
		// Set AuthHandler callback
		// This is called everytime a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
			if key, ok := usersMap[username]; ok {
				return key, true
			}
			return nil, false
		},
		// ListenerConfig is a list of Listeners and the configuration around them
		ListenerConfigs: []turn.ListenerConfig{
			{
				Listener:              tcpListener,
				RelayAddressGenerator: &I2PRelayAddressGenerator{},
			},
		},
	})
	if err != nil {
		log.Panic(err)
	}

	// Block until user sends SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if err = s.Close(); err != nil {
		log.Panic(err)
	}
}
