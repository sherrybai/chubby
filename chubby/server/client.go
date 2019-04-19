package server

import (
	"log"
	"github.com/hashicorp/raft"
	"net/rpc"
)

type LockName string

type LockPath string

type LockClient struct {
    /* Client transport layer. */
    trans           *raft.NetworkTransport
    /* Location of master server. */
    masterServer   raft.ServerAddress
    /* Location of locks. */
    locks           map[LockName]LockPath
    /* Session */
    session 		*Session
}

/* Create lock client. */
func CreateLockClient(trans *raft.NetworkTransport, masterServer raft.ServerAddress) (*LockClient, error) {
    lc := &LockClient {
        trans:          trans,
        masterServer:   masterServer,
        locks:          make(map[LockName]LockPath),
    }
    return lc, nil
}

func CreateLock(lc *LockClient, name LockName, path LockPath) error {
	// Set up TCP connection
	client, err := rpc.Dial("tcp", string(lc.masterServer))
	checkError(err)
	args :=  CreateLockRequest{name: name, path: path}
	response := clientResponse{false}
	err = client.Call("Handler.Create", args, &response)
	if err != nil {
		return err
	}
	return nil
}

func DeleteLock(lc *LockClient, name LockName, path LockPath) error {
	client, err := rpc.Dial("tcp", string(lc.masterServer))
	checkError(err)
	args :=  CreateLockRequest{name: name, path: path}
	response := clientResponse{false}
	err = client.Call("Handler.Create", args, &response)
	if err != nil {
		return err
	}
	return nil
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


