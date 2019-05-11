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
)

var client_fast_reqs_id		string		// ID of this client.

func init() {
	flag.StringVar(&client_fast_reqs_id, "clientID", "client_fast_reqs", "ID of this client")
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


	shouldAcquire := true
	for {
		select {
		case <- quitCh:
			fmt.Println("Exiting.")
			return

		default:
			if shouldAcquire {
				ok, err := sess.TryAcquireLock(lockName, api.SHARED)
				if err != nil {
					continue
				}
				if !ok {
					fmt.Println("Failed to acquire lock. Continuing.")
					continue
				}
				shouldAcquire = false
			}

			if !shouldAcquire {
				err = sess.ReleaseLock(lockName)
				if err != nil {
					fmt.Printf("Release failed with error: %s\n", err.Error())
					continue
				}
				shouldAcquire = true
			}
		}
	}
}
