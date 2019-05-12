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

func acquire_release(clientID string, sess *client.ClientSession) {
	var lockNames []string
	for i := 0; i < 100; i++ {
		lockName := fmt.Sprintf("lock_%d_%s",i, string(clientID))
		err := sess.OpenLock(api.FilePath(lockName))
		if err != nil {
			fmt.Println("Failed to open lock. Exiting.")
		}
		lockNames = append(lockNames, lockName)
	}
	for i := 0; i < 1000000; i++ {
		for j := 0; j < 100; j++ {
			ok, err := sess.TryAcquireLock(api.FilePath(lockNames[j]), api.EXCLUSIVE)
			if !ok {
				log.Println("Failed to acquire lock. Continuing.")
			}
			if err != nil {
				log.Fatal(err)
			}
			err = sess.ReleaseLock(api.FilePath(lockNames[j]))

			if err != nil {
				log.Printf("Release failed with error: %s\n", err.Error())
				continue
			}
		}
	}
}

func main () {
	var clientIDs []string
	var sessions []*client.ClientSession
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for i := 0; i < 200; i++ {
		clientIDs = append(clientIDs, fmt.Sprintf("client_%d",i))
	}
	for i := 0; i < 200; i++ {
		sess,err := client.InitSession(api.ClientID(clientIDs[i]))
		if err != nil {
			log.Fatal(err)
		}
		sessions = append(sessions, sess)
	}
	for i :=0; i < 199; i++ {
		go acquire_release(clientIDs[i], sessions[i])
	}

	var lockNames []string
	for i := 0; i < 100; i++ {
		lockName := fmt.Sprintf("lock_%d_%s",i, string(clientIDs[99]))
		err := sessions[199].OpenLock(api.FilePath(lockName))
		if err != nil {
			fmt.Println("Failed to open lock. Exiting.")
		}
		lockNames = append(lockNames, lockName)
	}

	counter := 0
	go func(count *int) {
		for  range time.Tick(time.Second) {
			fmt.Printf("We have performed %d operations in the last second\n", *count)
			*count = 0
		}
	}(&counter)

	for i := 0; i < 1000000; i++ {
		for j := 0; j < 100; j++ {
			ok, err := sessions[199].TryAcquireLock(api.FilePath(lockNames[j]), api.EXCLUSIVE)
			if !ok {
				log.Println("Failed to acquire lock. Continuing.")
			}
			if err != nil {
				log.Fatal(err)
			}
			err = sessions[199].ReleaseLock(api.FilePath(lockNames[j]))

			if err != nil {
				log.Printf("Release failed with error: %s\n", err.Error())
				continue
			}
			counter += 1
		}
	}


	// Exit on signal.
	<-quitCh

}
