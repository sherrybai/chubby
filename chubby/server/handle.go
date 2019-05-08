// Define RPC calls accepted by Chubby server.

package server

import "errors"

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
	error string
}


type KeepAliveRequest struct {
	params 		[]byte
}

type KeepAliveResponse struct {
	response	[]byte
}

type createSessionRequest struct {
	clientID 		ClientID
}

type createSessionResponse struct {
	error string
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
	// Maybe here: create a new thread for each session that handles session timeouts.
	// Put some new method "handleSession" in server.go
	// then call go handleSession()
	// -> handleSession can check if timeout has expired at regular intervals?
	sess, _ := CreateSession(clientID ClientID)


}

// KeepAlive calls allow the client to extend the Chubby session.
func (h *Handler) KeepAlive(req ClientRequest, res *ClientResponse) error {

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
