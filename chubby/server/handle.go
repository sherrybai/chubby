// Define RPC calls accepted by Chubby server.

package server

import (
	"errors"
	"fmt"
	"cos518project/chubby/api"
)

/*
 * Additional RPC interfaces available only to servers.
 */

type JoinRequest struct {
	RaftAddr string
	NodeID string
}

type JoinResponse struct {
	Error error
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
func (h *Handler) InitSession(req api.InitSessionRequest, res *api.InitSessionResponse) error {
	//if app.address != string(app.store.Raft.Leader()) {
	//	res.LeaderAddress = string(app.store.Raft.Leader())
	//	return nil
	//}
	_, err := CreateSession(api.ClientID(req.ClientID))
	if err != nil {
		return err
	}
	res.LeaderAddress = app.address
	return nil
}

// KeepAlive calls allow the client to extend the Chubby session.
func (h *Handler) KeepAlive(req api.KeepAliveRequest, res *api.KeepAliveResponse) error {
	//// If a non-leader node receives a KeepAlive, return error
	//if app.address != string(app.store.Raft.Leader()) {
	//	return errors.New(fmt.Sprintf("Node %s is not the leader", app.address))
	//}

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
func (h *Handler) OpenLock(req api.OpenLockRequest, res *api.OpenLockResponse) error {
	sess, ok := app.sessions[req.ClientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.ClientID))
	}
	err := sess.OpenLock(req.Filepath)
	if err != nil {
		return err
	}
	return nil
}

// Delete a lock.
func (h *Handler) DeleteLock(req api.DeleteLockRequest, res *api.DeleteLockResponse) error {
	sess, ok := app.sessions[req.ClientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.ClientID))
	}
	err := sess.DeleteLock(req.Filepath)
	if err != nil {
		return err
	}
	return nil
}

// Try to acquire a lock.
func (h *Handler) TryAcquireLock(req api.TryAcquireLockRequest, res *api.TryAcquireLockResponse) error {
	sess, ok := app.sessions[req.ClientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.ClientID))
	}
	isSuccessful, err := sess.TryAcquireLock(req.Filepath, req.Mode)
	if err != nil {
		return err
	}
	res.IsSuccessful = isSuccessful
	return nil
}

// Release lock.
func (h *Handler) ReleaseLock(req api.ReleaseLockRequest, res *api.ReleaseLockResponse) error {
	sess, ok := app.sessions[req.ClientID]
	if !ok {
		return errors.New(fmt.Sprintf("No session exists for %s", req.ClientID))
	}
	err := sess.ReleaseLock(req.Filepath)
	if err != nil {
		return err
	}
	return nil
}