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

var clientID1		string		// ID of client 1
var clientID2		string		// ID of client 2

func init() {
	flag.StringVar(&clientID1, "clientID1", "simple_client_1", "ID of client 1")
	flag.StringVar(&clientID2, "clientID2", "simple_client_2", "ID of client 2")
}

func main() {
	// Parse flags from command line.
	flag.Parse()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sess1, err := client.InitSession(api.ClientID(clientID1))
	sess2, err := client.InitSession(api.ClientID(clientID2))
	errOpenLock1 := sess1.OpenLock("LOCK/Lock1")
	errOpenLock2 := sess2.OpenLock("LOCK/Lock2")
	if errOpenLock1 != nil {
		log.Printf("Session 1 has trouble opening lock")
		log.Fatal(errOpenLock1)
	}
	if errOpenLock2 != nil {
		log.Printf("Session 2 has trouble opening lock")
		log.Fatal(errOpenLock2)
	}

	if err != nil {
		log.Fatal(err)
	}

	// Exit on signal.
	<-quitCh
}
