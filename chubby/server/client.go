package server

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net/rpc"
	"os"
	"strings"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

type LockName string

type LockPath string

type LockClient struct {
    /* Client transport layer. */
    trans           *raft.NetworkTransport
    /* Location of master server. */
    masterServer   raft.ServerAddress
    /* Location of locks. */
    locks           map[LockPath]LockName
    /* Session */
    session 		*session.Session
}

/* Create lock client. */
func CreateLockClient(trans *raft.NetworkTransport, masterServer raft.ServerAddress) (*LockClient, error) {
    lc := &LockClient {
        trans:          trans,
        masterServers:  masterServer,
        locks:          make(map[LockName]LockPath),
    }
    return lc, nil
}

func (lc *LockClient) CreateLock(name LockName, path LockPath) (error) {
	client, err: = rpc.Dial("tcp", lc.masterServer)
	checkError(err)
	args :=  CreastLockRequest{name, path}
	response := clientResponse{false}
	err = client.call("Handler.Create", args, &response)
	if err != nil {
		return err
	}
	return nil
}

func (lc *LockClient) DeleteLock (path LockPath) (error) {
	client, err: = rpc.Dial("tcp", lc.masterServer)
	checkError(err)
	args :=  clientRequest{path}
	response := clientResponse{false}
	err = client.call("Handler.Delete", args, &response)
	if err != nil {
		return err
	}
	return nil
}

func (lc *LockClient) AcquireLock (path LockPath) (error) {
	client, err: = rpc.Dial("tcp", lc.masterServer)
	checkError(err)
	args :=  clientRequest{path}
	response := acquireLockResponse{false, ""}
	err = client.call("Handler.Acquire", args, &response)
	if err != nil {
		return err
	}
	if response.isSuccessful == true {
		lc.loc
	}
	return nil
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


