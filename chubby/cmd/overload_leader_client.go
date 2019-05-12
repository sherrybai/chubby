package main

import (
	"cos518project/chubby/api"
	"cos518project/chubby/client"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("Start Time is %s\n", start.String())
	fmt.Printf("%s latency: %s\n", name, elapsed)
}

func main () {
	var clientIDs []string
	var sessions []*client.ClientSession
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for i := 0; i < 100; i++ {
		clientIDs = append(clientIDs, fmt.Sprintf("client_%d",i))
	}
	for i := 0; i < 100; i++ {
		sess,err := client.InitSession(api.ClientID(clientIDs[i]))
		if err != nil {
			log.Fatal(err)
		}
		sessions = append(sessions, sess)
	}
	// Exit on signal.
	<-quitCh

}
