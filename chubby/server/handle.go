// Define RPC calls accepted by Chubby server.

package server

import (
	"cos518project/chubby/api"
	"errors"
	"fmt"
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
	// If a non-leader node receives an InitSession, return error
	if app.store.RaftBind != string(app.store.Raft.Leader()) {
		return errors.New(fmt.Sprintf("Node %s is not the leader", app.address))
	}

	_, err := CreateSession(api.ClientID(req.ClientID))
	return err
}

// KeepAlive calls allow the client to extend the Chubby session.
func (h *Handler) KeepAlive(req api.KeepAliveRequest, res *api.KeepAliveResponse) error {
	// If a non-leader node receives a KeepAlive, return error
	if app.store.RaftBind != string(app.store.Raft.Leader()) {
		return errors.New(fmt.Sprintf("Node %s is not the leader", app.address))
	}

	var err error
	sess, ok := app.sessions[req.ClientID]
	if !ok {
		// Probably a jeopardy KeepAlive: create a new session for the client
		app.logger.Printf("Client %s sent jeopardy KeepAlive: creating new session", req.ClientID)

		// Note: this starts the local lease countdown
		// Should be ok to not call KeepAlive until later because lease TTL is pretty long (12s)
		sess, err = CreateSession(req.ClientID)
		if err != nil {
			// This shouldn't happen because session shouldn't be in app.sessions struct yet
			return err
		}

		app.logger.Printf("New session for client %s created", req.ClientID)

		// For each lock in the KeepAlive, try to acquire the lock
		// If any of the acquires fail, terminate the session.
		for filePath, lockMode := range(req.Locks) {
			ok, err := sess.TryAcquireLock(filePath, lockMode)
			if err != nil {
				// Don't return an error because the session won't terminate!
				app.logger.Printf("Error when client %s acquiring lock at %s: %s", req.ClientID, filePath, err.Error())
			}
			if !ok {
				app.logger.Printf("Jeopardy client %s failed to acquire lock at %s", req.ClientID, filePath)
				// This should cause the KeepAlive response to return that session should end.
				sess.TerminateSession()

				return nil // Don't return an error because the session won't terminate!
			}
			app.logger.Printf("Lock %s reacquired successfully.", filePath)
		}

		app.logger.Printf("Finished jeopardy KeepAlive process for client %s", req.ClientID)
	}

	duration := sess.KeepAlive(req.ClientID)
	res.LeaseLength = duration
	return nil
}

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