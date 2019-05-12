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
/*func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("Start Time is %s\n", start.String())
	fmt.Printf("%s latency: %s\n", name, elapsed)
}*/

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

	//var startTime time.Time
	counter := 0
	go func(count *int) {
		for  range time.Tick(time.Second) {
			fmt.Printf("We have performed %d operations in the last second\n", *count)
			*count = 0
		}
	}(&counter)
	for {
		select {
		case <- quitCh:
			fmt.Println("Exiting.")
			return

		default:
			//startTime = time.Now()
			ok, err := sess.TryAcquireLock(lockName, api.SHARED)
			//timeTrack(startTime, "TryAcquire")

			if err != nil {
				if sess.IsExpired() {
					sess, err = client.InitSession(api.ClientID(client_fast_reqs_id))
					if err != nil {
						log.Fatal(err)
					} else {
						continue
					}
				}
				log.Printf("TryAcquire failed with error: %s\n", err.Error())
				// Say we acquire a lock, then we pause that node
				// Release will fail with an error saying connection is closed
				// When the node is backup before jeopardy is over, the client
				// would hold the lock, and if we don't comment out continue,
				// the client will just keep getting the client already owns this lock
				// error
				//continue
			}
			if !ok {
				log.Println("Failed to acquire lock. Continuing.")
				//continue
			}

			//startTime = time.Now()
			err = sess.ReleaseLock(lockName)
			//timeTrack(startTime, "Release")

			if err != nil {
				log.Printf("Release failed with error: %s\n", err.Error())
				continue
			}
			counter += 1
		}
	}
}
