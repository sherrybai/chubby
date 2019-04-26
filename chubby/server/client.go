package server

import (
	"errors"
	"fmt"
	"encoding/json"
)

// Mode of a lock
type LockMode	int
const (
	EXCLUSIVE LockMode = iota
	SHARED
)

// Session contains metadata for one Chubby session.
// For simplicity, we say that each client can only init one session with
// the Chubby servers.
type Session struct {
	// Client to which this Session corresponds.
	clientID 		ClientID

    // A data structure describing which locks the client holds.
    // Maps lock filepath -> Lock struct.
    locks           map[FilePath]*Lock
}

// Lock describes information about a particular Chubby lock.
type Lock struct {
	path			FilePath  // The path to this lock in the store.
	mode			LockMode  // Shared or exclusive lock?
	owners			map[ClientID]bool  // Who is holding the lock?
}

//// Server checks against this metadata when clients perform operations
//// on file handles.
//type Handle struct {
////	// Sequence number.
////	// Allows us to check if this handle was created by a previous master.
////	seq			int
////
//	clientID   	ClientID   // Client to which this handle corresponds.
//	path		FilePath   // Path of lock to which this handle corresponds.
//}

/* Create Session struct. */
func CreateSession(clientID ClientID) (*Session, error) {
    sess := &Session{
        clientID:   	clientID,
        locks:          make(map[FilePath]*Lock),
    }
    return sess, nil
}

// Create the lock if it does not exist.
// Add the lock to list of this session's open files.
func Open(clientID ClientID, path FilePath) error {
	// Check if lock exists in persistent store
	_, err := app.store.Get(string(path))
	if err != nil {
		// Add lock to persistent store: (key: LockPath, value: "")
		err = app.store.Set(string(path), "")
		if err != nil {
			return err
		}
	}



	return nil
}

func (sess *Session) DeleteLock(path FilePath) error {
	// If we are not holding the lock, we cannot delete it.
	lock, exists := sess.locks[path]
	if !exists {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s", path))
	}

	// Check if we are holding the lock in exclusive mode
	if lock.mode != EXCLUSIVE {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s in exclusive mode", path))
	}

	// Delete the lock from Session metadata.
	delete(sess.locks, path)

	// Delete the lock from the store.
	err := app.store.Delete(string(path))
	return err
}

func (sess *Session) Try_AcquireLock (path FilePath, mode LockMode) (error) {
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
			sess.sessionID: true,
		}
		err = app.store.Set(path, json.Marshal(&lock))
		/* More for debugging purpose, might need to do something else late for this error*/
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to Commit in Store"))
		}
		sess.locks[path] = lock
	} else {
		if lock.mode == EXCLUSIVE {
			return errors.New(fmt.Sprintf("The lock is being held in exclusive mode."))
		}
		lock.mode = mode
		lock.owners[sess.sessionID] = true
		err = app.store.Set(path, json.Marshal(&lock))
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to Commit in Store"))
		}
		sess.locks[path] = lock
	}
	return nil
}

func (sess *Session) ReleaseLock (path FilePath) (error) {
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
		sess.locks[path] = false
		delete(lock.owners, sess.sessionID)
		err = app.store.Set(path, json.Marshal(&lock))
		if err != nil {
			return errors.New(fmt.Sprintf("Fail to commit"))
		}
	} else if lock.mode == SHARED {
		if len(lock.owners) == 1 {
			lock.mode = FREE
			sess.locks[path] = false
			delete(lock.owners, sess.sessionID)
			err = app.store.Set(path, json.Marshal(&lock))
			if err != nil {
				return errors.New(fmt.Sprintf("Fail to commit"))
			}
		} else {
			sess.locks[path] = false
			delete(lock.owners, sess.sessionID)
			err = app.store.Set(path, json.Marshal(&lock))
			if err != nil {
				return errors.New(fmt.Sprintf("Fail to commit"))
			}
		}
	}
	return nil
}
