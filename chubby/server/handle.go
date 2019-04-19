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

type CreateLockRequest struct {
	name LockName
	path LockPath
}

type clientRequest struct {
	path LockPath
}

type clientResponse struct {
	isSuccessful bool
}

type EmptyResponse struct {}

// RPC handler type
type Handler int

// Join the caller server to our server.
func (h *Handler) Join(req JoinRequest, res *EmptyResponse) error {
	
	return app.store.Join(req.NodeID, req.RaftAddr)
}

func (h *Handler) Create(req clientRequest, res *clientResponse) error

func (h *Handler) Delete(req clientRequest, res *clientResponse) error

func (h *Handler) Acquire(req clientRequest, res *clientResponse) error

func (h *Handler) Release(req clientRequest, res *clientResponse) error
// Open the file handler
func (h *Handler) Open() error {
	return 
}