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

type EmptyResponse struct {}


type ClientRequest struct {
	params 		[]byte
}

type ClientResponse struct {
	response	[]byte
}

// RPC handler type
type Handler int

/*
 * Called by servers:
 */

// Join the caller server to our server.
func (h *Handler) Join(req JoinRequest, res *EmptyResponse) error {
	return app.store.Join(req.NodeID, req.RaftAddr)
}

/*
 * Called by clients:
 */

// Create a lock.
func (h *Handler) Create(req ClientRequest, res *ClientResponse) error {

}

// Delete a lock.
func (h *Handler) Delete(req ClientRequest, res *ClientResponse) error {

}

// Acquire lock.
func (h *Handler) Acquire(req ClientRequest, res *ClientResponse) error {

}

// Release lock.
func (h *Handler) Release(req ClientRequest, res *ClientResponse) error {

}
