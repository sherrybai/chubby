// Define RPC calls accepted by Chubby server.

package server

import (
	"encoding/json"
	"time"
)

type JoinRequest struct {
	RaftAddr string
	NodeID string
}

type JoinResponse struct {
	error error
}


type initSessionRequest struct {
	clientID ClientID
}

type initSessionResponse struct {
	leaderAddress string
	leaseLength time.Duration
}

type keepAliveRequest struct {
	clientID ClientID
}

type keepAliveResponse struct {
	leaseLength time.Duration
}

type Response struct {
	leaderAddress string
	leaseLength time.Duration
}

type Request struct {
	mode LockMode
	path FilePath
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
func (h *Handler) InitSession(req initSessionRequest, res *initSessionResponse) error {
	// If we are not the leader, return address of actual leader
	if app.address != string(app.store.Raft.Leader()) {
		res.leaseLength = 0
		res.leaderAddress = string(app.store.Raft.Leader())
	}

	// Create a new session.
	sess, err := CreateSession(ClientID(req.clientID))
	if err != nil {
		return err
	}

	// Respond with address of current leader.
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

//// Open a lock.
//func (h *Handler) OpenLock(req ClientRequest, res *ClientResponse) error {
//
//}
//
//// Delete a lock.
//func (h *Handler) DeleteLock(req ClientRequest, res *ClientResponse) error {
//
//}
//
//// Try to acquire a lock.
//func (h *Handler) TryAcquireLock(req ClientRequest, res *ClientResponse) error {
//
//}
//
//// Release lock.
//func (h *Handler) ReleaseLock(req ClientRequest, res *ClientResponse) error {
//
//}
