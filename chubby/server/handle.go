// Define RPC calls accepted by Chubby server.

package server

import (
	"errors"
	"fmt"
	"time"
)

/*
 * RPC interfaces.
 */

type JoinRequest struct {
	RaftAddr string
	NodeID string
}

type JoinResponse struct {
	Error error
}

type InitSessionRequest struct {
	ClientID ClientID
}

type InitSessionResponse struct {
	LeaderAddress string
}

type KeepAliveRequest struct {
	ClientID ClientID
}

type KeepAliveResponse struct {
	LeaseLength time.Duration
}

// TODO: make all fields exported

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
	res.Error = err
	return err
}

/*
 * Called by clients:
 */

// Initialize a client-server session.
func (h *Handler) InitSession(req InitSessionRequest, res *InitSessionResponse) error {
	if app.address != string(app.store.Raft.Leader()) {
		res.LeaderAddress = string(app.store.Raft.Leader())
		return nil
	}
	_, err := CreateSession(ClientID(req.ClientID))
	if err != nil {
		return err
	}
	res.LeaderAddress = app.address
	return nil
}

// KeepAlive calls allow the client to extend the Chubby session.
func (h *Handler) KeepAlive(req KeepAliveRequest, res *KeepAliveResponse) error {
	// If a non-leader node receives a KeepAlive, return error
	if app.address != string(app.store.Raft.Leader()) {
		return errors.New(fmt.Sprintf("Node %s is not the leader", app.address))
	}

	// TODO: change this to handle failovers
	sess, ok := app.sessions[req.ClientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.ClientID))
	}

	duration, err := sess.KeepAlive(req.ClientID)
	if err != nil {
		return err
	}
	res.LeaseLength = duration
	return nil
}

// Chubby API methods for handling locks.
// Each method corresponds to a method in session.go.
// Open a lock.
func (h *Handler) OpenLock(req OpenLockRequest, res *OpenLockResponse) error {
	sess, ok := app.sessions[req.clientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.clientID))
	}
	err := sess.OpenLock(req.filepath)
	if err != nil {
		return err
	}
	return nil
}

// Delete a lock.
func (h *Handler) DeleteLock(req DeleteLockRequest, res *DeleteLockResponse) error {
	sess, ok := app.sessions[req.clientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.clientID))
	}
	err := sess.DeleteLock(req.filepath)
	if err != nil {
		return err
	}
	return nil
}

// Try to acquire a lock.
func (h *Handler) TryAcquireLock(req TryAcquireLockRequest, res *TryAcquireLockResponse) error {
	sess, ok := app.sessions[req.clientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.clientID))
	}
	isSuccessful, err := sess.TryAcquireLock(req.filepath, req.mode)
	if err != nil {
		return err
	}
	res.isSuccessful = isSuccessful
	return nil
}

// Release lock.
func (h *Handler) ReleaseLock(req ReleaseLockRequest, res *ReleaseLockResponse) error {
	sess, ok := app.sessions[req.clientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.clientID))
	}
	err := sess.ReleaseLock(req.filepath)
	if err != nil {
		return err
	}
	return nil
}