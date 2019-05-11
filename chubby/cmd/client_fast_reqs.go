// This client continuously acquires and releases locks.

package main

import (
	"cos518project/chubby/api"
	"cos518project/chubby/client"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var client_fast_reqs_id		string		// ID of this client.

func init() {
	flag.StringVar(&client_fast_reqs_id, "clientID", "client_fast_reqs", "ID of this client")
}

// Adapted from: https://coderwall.com/p/cp5fya/measuring-execution-time-in-go
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("Start Time is %s\n", start.Second())
	fmt.Printf("%s latency: %s\n", name, elapsed)
}

func main() {
	// Parse flags from command line.
	flag.Parse()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sess, err := client.InitSession(api.ClientID(client_fast_reqs_id))

	if err != nil {
		log.Fatal(err)
	}

	var lockName api.FilePath = "lock"

	err = sess.OpenLock(lockName)
	if err != nil {
		fmt.Println("Failed to open lock. Exiting.")
	}

	fmt.Println("Begin acquiring and releasing locks.")

	var startTime time.Time

	for {
		select {
		case <- quitCh:
			fmt.Println("Exiting.")
			return

		default:
			startTime = time.Now()
			ok, err := sess.TryAcquireLock(lockName, api.SHARED)
			timeTrack(startTime, "TryAcquire")

			if err != nil {
				log.Printf("TryAcquire failed with error: %s\n", err.Error())
				continue
			}
			if !ok {
				log.Println("Failed to acquire lock. Continuing.")
				continue
			}

			startTime = time.Now()
			err = sess.ReleaseLock(lockName)
			timeTrack(startTime, "Release")

			if err != nil {
				log.Printf("Release failed with error: %s\n", err.Error())
				continue
			}
		}
	}
}
