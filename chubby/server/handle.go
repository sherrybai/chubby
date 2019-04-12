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

// RPC handler type
type Handler int

// Join the caller server to our server.
func (h *Handler) Join(req JoinRequest, res *EmptyResponse) error {
	
	return app.store.Join(req.NodeID, req.RaftAddr)
}

func (h *Handler) Create(req CreateRequest, res *EmptyResponse) error

func (h *Handler) Delete(req CreateRequest, res *EmptyResponse) error

func (h *Handler) Acquire(req AcquireRequest, res *EmptyResponse) error

func (h *Handler) Release(req ReleaseRequest, res *EmptyResponse) error
// Open the file handler
func (h *Handler) Open() error {
	return 
}