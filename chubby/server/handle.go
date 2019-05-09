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


type InitSessionRequest struct {
	clientID ClientID
}

type InitSessionResponse struct {
	leaderAddress string
	isLeader bool
}

type KeepAliveRequest struct {
	clientID ClientID
}

type KeepAliveResponse struct {
	leaseLength time.Duration
}

type OpenLockRequest struct {
	clientID ClientID
	filepath FilePath
}

type OpenLockResponse struct {

}

type DeleteLockRequest struct {
	clientID ClientID
	filepath FilePath
}

type DeleteLockResponse struct {

}

type TryAcquireLockRequest struct {
	clientID ClientID
	filepath FilePath
	mode LockMode
}

type TryAcquireLockResponse struct {
	isSuccessful bool
}

type ReleaseLockRequest struct {
	clientID ClientID
	filepath FilePath
}

type ReleaseLockResponse struct {

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
func (h *Handler) InitSession(req InitSessionRequest, res *InitSessionResponse) error {
	if app.address != string(app.store.Raft.Leader()) {
		res.isLeader = false
		res.leaderAddress = string(app.store.Raft.Leader())
		return nil

	}
	_, err := CreateSession(ClientID(req.clientID))
	if err != nil {
		return err
	}
	res.isLeader = true
	res.leaderAddress = string(app.store.Raft.Leader())
	return nil
}

// KeepAlive calls allow the client to extend the Chubby session.
func (h *Handler) HandleKeepAlive(req KeepAliveRequest, res *KeepAliveResponse) error {
	session := app.sessions[ClientID(req.clientID)]
	duration, err := session.KeepAlive(ClientID(req.clientID))
	if err != nil {
		return err
	}
	res.leaseLength = duration
	return nil
}

// Chubby API methods for handling locks.
// Each method corresponds to a method in session.go.

// Open a lock.
func (h *Handler) OpenLock(req OpenLockRequest, res *OpenLockResponse) error {
	session := app.sessions[ClientID(req.clientID)]
	err := session.OpenLock(req.filepath)
	if err != nil {
		return err
	}
	return nil
}

// Delete a lock.
func (h *Handler) DeleteLock(req DeleteLockRequest, res *DeleteLockResponse) error {
	session := app.sessions[ClientID(req.clientID)]
	err := session.DeleteLock(req.filepath)
	if err != nil {
		return err
	}
	return nil
}

// Try to acquire a lock.
func (h *Handler) TryAcquireLock(req TryAcquireLockRequest, res *TryAcquireLockResponse) error {
	session := app.sessions[ClientID(req.clientID)]
	isSuccessful, err := session.TryAcquireLock(req.filepath, req.mode)
	if err != nil {
		return err
	}
	res.isSuccessful = isSuccessful
	return nil
}

// Release lock.
func (h *Handler) ReleaseLock(req ReleaseLockRequest, res *ReleaseLockResponse) error {
	session := app.sessions[ClientID(req.clientID)]
	err := session.ReleaseLock(req.filepath)
	if err != nil {
		return err
	}
	return nil
}
