// This client does nothing but maintain a session with the Chubby server.

package main

import (
	"cos518project/chubby/api"
	"cos518project/chubby/client"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var clientID		string		// ID of this client.

func init() {
	flag.StringVar(&clientID, "clientID", "simple_client_1", "ID of this client")
}

func main() {
	// Parse flags from command line.
	flag.Parse()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	_, err := client.InitSession(api.ClientID(clientID))


	if err != nil {
		log.Fatal(err)
	}

	// Exit on signal.
	<-quitCh
}
