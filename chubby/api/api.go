// Interfaces for RPC calls between clients and servers.

package api

import "time"

/*
 * Shared types.
 */

type ClientID string
type FilePath 	string

// Mode of a lock
type LockMode	int
const (
	EXCLUSIVE LockMode = iota
	SHARED
	FREE
)

/*
 * RPC interfaces.
 */

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
	ClientID ClientID
	Filepath FilePath
}

type OpenLockResponse struct {

}

type DeleteLockRequest struct {
	ClientID ClientID
	Filepath FilePath
}

type DeleteLockResponse struct {

}

type TryAcquireLockRequest struct {
	ClientID ClientID
	Filepath FilePath
	Mode LockMode
}

type TryAcquireLockResponse struct {
	IsSuccessful bool
}

type ReleaseLockRequest struct {
	ClientID ClientID
	Filepath FilePath
}

type ReleaseLockResponse struct {

}