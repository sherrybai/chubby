// Define RPC calls accepted by Chubby server.

package server

import (
	"errors"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

var (
	ErrParams        = errors.New("ERR params invalid")
	ErrRespType      = errors.New("ERR resp type invalid")
	ErrCmdNotSupport = errors.New("ERR command not supported")
)

type JoinRequest struct {
	RaftAddr string
	NodeID string
}

type JoinResponse struct {
	error error
}


type ClientRequest struct {
	clientID	string
	params 		[]byte
}

type ClientResponse struct {
	response	[]byte
}

type Response struct {
	leaderAddress string
	leaseLength time.Duration
}

// RPC handler type
type Handler int

/*
 * Called by servers:
 */

// Join the caller server to our server.
func (h *Handler) Join(req JoinRequest, res *JoinResponse) error {
	err := app.store.Join(req.NodeID, req.RaftAddr)
	res.error = err
	return err
}

/*
 * Called by clients:
 */

// Initialize a client-server session.
func (h *Handler) InitSession(req ClientRequest, res *ClientResponse) error {
	if app.address != string(app.store.Raft.Leader()) {
		response := &Response{leaderAddress: string(app.store.Raft.Leader()), leaseLength: 0*time.Second}
		b, err := json.Marshal(response)
		if err != nil {
			res.response = b
		} else {
			return err
		}
	}
	sess, err := CreateSession(ClientID(req.clientID))
	if err != nil {
		return nil
	}
	response := &Response{leaderAddress: string(app.store.Raft.Leader()), leaseLength: sess.leaseLength}
	b, err := json.Marshal(response)
	if err != nil {
		res.response = b
	} else {
		return err
	}
	return nil
}

// KeepAlive calls allow the client to extend the Chubby session.
func (h *Handler) HandleKeepAlive(req ClientRequest, res *ClientResponse) error {
	session := app.sessions[ClientID(req.clientID)]
	duration, _ := session.KeepAlive(ClientID(req.clientID))
	response := &Response{leaderAddress: string(app.store.Raft.Leader()), leaseLength: duration}
	b, err := json.Marshal(response)
	if err != nil {
		res.response = b
	} else {
		return err
	}
	return nil
}

// Chubby API methods for handling locks.
// Each method corresponds to a method in session.go.

// Open a lock.
func (h *Handler) OpenLock(req ClientRequest, res *ClientResponse) error {

}

// Delete a lock.
func (h *Handler) DeleteLock(req ClientRequest, res *ClientResponse) error {

}

// Try to acquire a lock.
func (h *Handler) TryAcquireLock(req ClientRequest, res *ClientResponse) error {

}

// Release lock.
func (h *Handler) ReleaseLock(req ClientRequest, res *ClientResponse) error {

}
