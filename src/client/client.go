package blindclient

import (
	//"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/eyedeekay/sam3"
	//"github.com/eyedeekay/sam3/i2pkeys"
	"github.com/pion/logging"
	"github.com/pion/turn/v2"
)

func Main(host, user, realm string, ping bool) {
	if len(host) == 0 {
		log.Fatalf("'host' is required")
	}

	if len(user) == 0 {
		log.Fatalf("'user' is required")
	}

	// Dial TURN Server
	//conn, err := net.Dial("tcp", host)
	sam, err := sam3.NewSAM("127.0.0.1:7657")
	if err != nil {
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		return
	}
	stream, err := sam.NewStreamSession(keys.Addr().Base32()[0:9], keys, sam3.Options_Small)
	if err != nil {
		return
	}
	//fmt.Println("Client: Connecting to " + server.Base32())
	conn, err := stream.Dial("i2p", host)
	if err != nil {
		panic(err)
	}

	cred := strings.Split(user, "=")

	// Start a new TURN Client and wrap our net.Conn in a STUNConn
	// This allows us to simulate datagram based communication over a net.Conn
	cfg := &turn.ClientConfig{
		STUNServerAddr: host,
		TURNServerAddr: host,
		Conn:           turn.NewSTUNConn(conn),
		Username:       cred[0],
		Password:       cred[1],
		Realm:          realm,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Start listening on the conn provided.
	err = client.Listen()
	if err != nil {
		panic(err)
	}

	// Allocate a relay socket on the TURN server. On success, it
	// will return a net.PacketConn which represents the remote
	// socket.
	relayConn, err := client.Allocate()
	if err != nil {
		panic(err)
	}
	defer func() {
		if closeErr := relayConn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	// The relayConn's local address is actually the transport
	// address assigned on the TURN server.
	log.Printf("relayed-address=%s", relayConn.LocalAddr().String())

	// If you provided `-ping`, perform a ping test agaist the
	// relayConn we have just allocated.
	if ping {
		err = doPingTest(client, relayConn)
		if err != nil {
			panic(err)
		}
	}
}

func doPingTest(client *turn.Client, relayConn net.PacketConn) error {
	// Send BindingRequest to learn our external IP
	mappedAddr, err := client.SendBindingRequest()
	if err != nil {
		return err
	}

	// Set up pinger socket (pingerConn)
	//pingerConn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	sam, err := sam3.NewSAM("127.0.0.1:7657")
	if err != nil {
		return err
	}
	keys, err := sam.NewKeys()
	if err != nil {
		return err
	}
	pingerConn, err := sam.NewDatagramSession(keys.Addr().Base32()[0:9], keys, sam3.Options_Small, 0)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := pingerConn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	// Punch a UDP hole for the relayConn by sending a data to the mappedAddr.
	// This will trigger a TURN client to generate a permission request to the
	// TURN server. After this, packets from the IP address will be accepted by
	// the TURN server.
	_, err = relayConn.WriteTo([]byte("Hello"), mappedAddr)
	if err != nil {
		return err
	}

	// Start read-loop on pingerConn
	go func() {
		buf := make([]byte, 1500)
		for {
			n, from, pingerErr := pingerConn.ReadFrom(buf)
			if pingerErr != nil {
				break
			}

			msg := string(buf[:n])
			if sentAt, pingerErr := time.Parse(time.RFC3339Nano, msg); pingerErr == nil {
				rtt := time.Since(sentAt)
				log.Printf("%d bytes from from %s time=%d ms\n", n, from.String(), int(rtt.Seconds()*1000))
			}
		}
	}()

	// Start read-loop on relayConn
	go func() {
		buf := make([]byte, 1500)
		for {
			n, from, readerErr := relayConn.ReadFrom(buf)
			if readerErr != nil {
				break
			}

			// Echo back
			if _, readerErr = relayConn.WriteTo(buf[:n], from); readerErr != nil {
				break
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)

	// Send 10 packets from relayConn to the echo server
	for i := 0; i < 10; i++ {
		msg := time.Now().Format(time.RFC3339Nano)
		_, err = pingerConn.WriteTo([]byte(msg), relayConn.LocalAddr())
		if err != nil {
			return err
		}

		// For simplicity, this example does not wait for the pong (reply).
		// Instead, sleep 1 second.
		time.Sleep(time.Second)
	}

	return nil
}
