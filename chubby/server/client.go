package server

import (
	"errors"
	"fmt"
)

type SessionID 	string
type FilePath 	string

// Mode of a lock
type LockMode	int
const (
	EXCLUSIVE LockMode = iota
	SHARED
)

// LockClient describes the locks held by a particular client in a particular Chubby session.
type LockClient struct {
	// The session to which this LockClient corresponds.
	sessionID		SessionID

    // A data structure describing which locks this client holds.
    // Maps lock filepath -> Lock struct.
    locks           map[FilePath]*Lock
}

// Lock describes information about a particular Chubby lock.
type Lock struct {
	path			FilePath  // The path to this lock in the store.
	mode			LockMode  // Shared or exclusive lock?
	owners			map[string]bool  // Who is holding the lock?
}

/* Create lock client. */
func CreateLockClient(sessionID SessionID) (*LockClient, error) {
    lc := &LockClient {
        sessionID:   	sessionID,
        locks:          make(map[FilePath]*Lock),
    }
    return lc, nil
}

func (lc *LockClient) CreateLock(path FilePath, mode LockMode) error {
	// Check if lock exists in persistent store
	_, err := app.store.Get(string(path))
	if err == nil {
		return errors.New(fmt.Sprintf("Lock already exists at path %s", path))

	}
	// Add lock to persistent store: (key: LockPath, value: "")
	// TODO: What information should be in the lock file?
	err = app.store.Set(string(path), "")
	return err
}

func (lc *LockClient) DeleteLock (path FilePath) error {
	// If we are not holding the lock, we cannot delete it.
	lock, exists := lc.locks[path]
	if !exists {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s", path))
	}

	// Check if we are holding the lock in exclusive mode
	if lock.mode != EXCLUSIVE {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s in exclusive mode", path))
	}

	// Delete the lock from LockClient metadata.
	delete(lc.locks, path)

	// Delete the lock from the store.
	err := app.store.Delete(string(path))
	return err
}

func (lc *LockClient) AcquireLock (path FilePath) error {

}


