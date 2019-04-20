package server

import (
	"errors"
	"fmt"
	"encoding/json"
)

type SessionID 	string
type FilePath 	string

// Mode of a lock
type LockMode	int
const (
	EXCLUSIVE LockMode = 1
	SHARED LockMode = 2
	FREE LockMode = 3
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

func (lc *LockClient) Try_AcquireLock (path FilePath, mode LockMode) (error) {
	if mode != EXCLUSIVE && mode != SHARED {
		return errors.New(fmt.Sprintf("Invalid mode."))
	}
	lock_info, err := app.store.Get(path) 
	if err != nil {
		return errors.New(fmt.Sprintf("Lock at path %s doesn't exist", path))
	}
	lock_info_bytes = []byte(lock_info)
	var lock Lock
	err = json.Unmarshal(lock_info_bytes, &lock)
	if err != nil {
		return errors.New(fmt.Sprintf("Fail to Decode"))
	}
	if mode == EXCLUSIVE {
		if lock.mode != FREE {
			return errors.New(fmt.Sprintf("Lock is not free"))
		}
		lock.path = path
		lock.mode = mode
		lock.owners = map[string]bool {
			lc.sessionID: true,
		}
		err = app.store.Set(path, string(json.Marshal(&lock)))
		/* More for debugging purpose, might need to do something else late for this error*/
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to Commit in Store"))
		}
		lc.locks[path] = lock
	} else {
		if lock.mode == EXCLUSIVE {
			return errors.New(fmt.Sprintf("The lock is being held in exclusive mode."))
		}
		lock.mode = mode
		lock.owners[lc.sessionID] = true
		err = app.store.Set(path, string(json.Marshal(&lock)))
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to Commit in Store"))
		}
		lc.locks[path] = lock
	}
	return nil
}

func (lc *LockClient) ReleaseLock (path FilePath) (error) {
	lock_info, err := app.store.Get(path) 
	if err != nil {
		return errors.New(fmt.Sprintf("Lock at path %s doesn't exist", path))
	}
	lock_info_bytes = []byte(lock_info)
	var lock Lock
	err = json.Unmarshal(lock_info_bytes, &lock)
	if err != nil {
		return errors.New(fmt.Sprintf("Fail to Decode"))
	}
	if lock.mode == FREE {
		return nil
	} else if lock.mode == EXCLUSIVE {
		lock.mode = FREE
		lc.locks[path] = false
		delete(lock.owners, lc.sessionID)
		err = app.store.Set(path, string(json.Marshal(&lock)))
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to commit"))
		}
	} else if lock.mode == SHARED {
		if len(lock.owners) == 1 {
			lock.mode = FREE
			lc.locks[path] = false
			delete(lock.owners, lc.sessionID)
			err = app.store.Set(path, string(json.Marshal(&lock)))
			if err != nil {
				return errors.New(fmt.Sprintf("Fail to commit"))
			}
		} else {
			lc.locks[path] = false
			delete(lock.owners, lc.sessionID)
			err = app.store.Set(path, string(json.Marshal(&lock)))
			if err != nil {
				return errors.New(fmt.Sprintf("Fail to commit"))
			}
		}
	}
	return nil
}
