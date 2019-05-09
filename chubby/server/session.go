// Defines Chubby session metadata, as well as operations on locks that can
// be performed as part of a Chubby session.

package server

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Mode of a lock
type LockMode	int
const (
	EXCLUSIVE LockMode = iota
	SHARED
	FREE
)

const DefaultLeaseExt = 12 * time.Second

// Session contains metadata for one Chubby session.
// For simplicity, we say that each client can only init one session with
// the Chubby servers.
type Session struct {
	// Client to which this Session corresponds.
	clientID 		ClientID

	// Start time
	startTime		time.Time

	// Length of the Lease
	leaseLength     time.Duration

	//TTL Lock
	ttlLock 		sync.Mutex

	// Channel used to block Keepalive 
	ttlChannel  	chan string

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

/* Create Session struct. */
func CreateSession(clientID ClientID) (*Session, error) {
	if _, ok := app.sessions[clientID]; ok {
		return nil, errors.New(fmt.Sprintf("The client already has a session established with the master"))
	}

	// Create new session struct.
	sess := &Session{
        clientID:    clientID,
        startTime:   time.Now(),
        leaseLength: DefaultLeaseExt,
        ttlChannel:  make(chan string,1),
        locks:       make(map[FilePath]*Lock),
    }

	// Add the session to the sessions map.
	app.sessions[clientID] = sess

	// In a separate goroutine, periodically check if the lease is over
    go sess.MonitorSession()

    app.sessions[clientID] = sess
    return sess, nil
}

func (sess *Session) MonitorSession() {
	// At each second, check time until the lease is over.
	ticker := time.Tick(time.Second)
	for range ticker {
		durationLeaseOver := time.Until(sess.startTime.Add(leaseLength))
		if durationLeaseOver <= (2 * time.Second) {
			// Trigger KeepAlive response 2 seconds before timeout
			sess.ttlChannel <- "Ready"
		}
		if durationLeaseOver <= 0 {
			// Lease expired: destroy the session
			sess.DestroySession()
			return
		}
	}
}

// Destroy the session.
func (sess *Session) DestroySession() {
	delete(app.sessions, sess.clientID)

	app.logger.Printf("destroyed session with client %s", sess.clientID)
}

// Extend Lease after receiving keepalive messages
func (sess *Session) KeepAlive(clientID ClientID) (time.Duration, error) {
	if _, ok := app.sessions[sess.clientID]; !ok {
		return 0, errors.New(fmt.Sprintf("The current session is closed"))
	}
	// Block until shortly before lease expires
	<- sess.ttlChannel
	// Extend lease by 12 seconds
	sess.leaseLength = sess.leaseLength + DefaultLeaseExt

	app.logger.Printf("session with client %s extended: lease length %d", sess.clientID, sess.leaseLength)

	// Return new lease length.
	return sess.leaseLength, nil
}

// Create the lock if it does not exist.
func (sess *Session) OpenLock(path FilePath) error {
	if _, ok := app.sessions[sess.clientID]; !ok {
		return errors.New(fmt.Sprintf("The current session is closed"))
	}
	// Check if lock exists in persistent store
	_, err := app.store.Get(string(path))
	if err != nil {
		// Add lock to persistent store: (key: LockPath, value: "")
		err = app.store.Set(string(path), "")
		if err != nil {
			return err
		}

		// Add lock to in-memory struct of locks
		app.locks[path] = &Lock{
			path: path,
			mode: FREE,
			owners: make(map[ClientID]bool),
		}
	}

	return nil
}

// Delete the lock. Lock must be held in exclusive mode before calling DeleteLock.
func (sess *Session) DeleteLock(path FilePath) error {
	if _, ok := app.sessions[sess.clientID]; !ok {
		return errors.New(fmt.Sprintf("The current session is closed"))
	}
	// If we are not holding the lock, we cannot delete it.
	lock, exists := sess.locks[path]
	if !exists {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s", path))
	}

	// Check if we are holding the lock in exclusive mode
	if lock.mode != EXCLUSIVE {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s in exclusive mode", path))
	}

	// Check that the lock actually exists in the store.
	_, err := app.store.Get(string(path))

	if err != nil {
		return errors.New(fmt.Sprintf("Lock at %s does not exist in persistent store", path))
	}

	// Delete the lock from Session metadata.
	delete(sess.locks, path)

	// Delete the lock from in-memory struct of locks
	delete(app.locks, path)

	// Delete the lock from the store.
	err = app.store.Delete(string(path))
	return err
}

// Try to acquire the lock, returning either success (true) or failure (false).
func (sess *Session) TryAcquireLock (path FilePath, mode LockMode) (bool, error) {
	if _, ok := app.sessions[sess.clientID]; !ok {
		return false, errors.New(fmt.Sprintf("The current session is closed"))
	}
	// Validate mode of the lock.
	if mode != EXCLUSIVE && mode != SHARED {
		return false, errors.New(fmt.Sprintf("Invalid mode."))
	}

	// Do we already own the lock? Fail with error
	_, owned := sess.locks[path]
	if owned {
		return false, errors.New(fmt.Sprintf("We already own the lock at %s", path))
	}

	// Check if lock exists in persistent store
	_, err := app.store.Get(string(path))

	if err != nil {
		return false, errors.New(fmt.Sprintf("Lock at %s has not been opened", path))
	}

	// Check if lock exists in in-mem struct
	lock, exists := app.locks[path]
	if !exists {
		// Assume that some failure has occurred
		// Lazily recover lock struct: add lock to in-memory struct of locks
		// TODO: check if this is correct?
		app.locks[path] = &Lock{
			path: path,
			mode: FREE,
			owners: make(map[ClientID]bool),
		}
		lock = app.locks[path]
	}

	// Check the mode of the lock
	switch lock.mode {
	case EXCLUSIVE:
		// Should fail: someone probably already owns the lock
		if len(lock.owners) == 0 {
			// Throw an error if there are no owners but lock.mode is exclusive:
			// this means ReleaseLock was not implemented correctly
			return false, errors.New("Lock has EXCLUSIVE mode despite having no owners")
		} else if len(lock.owners) > 1 {
			return false, errors.New("Lock has EXCLUSIVE mode but has multiple owners")
		} else {
			// Fail with no error
			return false, nil
		}
	case SHARED:
		// If our mode is shared, then succeed; else fail
		if mode == EXCLUSIVE {
			return false, nil
		} else {  // mode == SHARED
			// Update lock owners
			lock.owners[sess.clientID] = true

			// Add lock to session lock struct
			sess.locks[path] = lock

			// Return success
			return true, nil
		}
	case FREE:
		// If lock has owners, either TryAcquireLock or ReleaseLock was not implemented correctly
		if len(lock.owners) > 0 {
			return false, errors.New("Lock has FREE mode but is owned by 1 or more clients")
		}

		// Should succeed regardless of mode
		// Update lock owners
		lock.owners[sess.clientID] = true

		// Update lock mode
		lock.mode = mode

		// Add lock to session lock struct
		sess.locks[path] = lock

		// Return success
		return true, nil
	default:
		return false, errors.New(fmt.Sprintf("Lock at %s has undefined mode %d", path, lock.mode))
	}
}

// Release the lock.
func (sess *Session) ReleaseLock (path FilePath) (error) {
	if _, ok := app.sessions[sess.clientID]; !ok {
		return errors.New(fmt.Sprintf("The current session is closed"))
	}
	// Check if lock exists in persistent store
	_, err := app.store.Get(string(path))

	if err != nil {
		return errors.New(fmt.Sprintf("Lock at %s does not exist in persistent store", path))
	}

	// Check if we own the lock.
	lock, present := sess.locks[path]

	// If not in session locks map, throw an error
	if !present || lock == nil {
		return errors.New(fmt.Sprintf("Lock at %s does not exist in session locks map", path))
	}

	// Check that we are among the owners of the lock.
	_, present = lock.owners[sess.clientID]
	if !present || !lock.owners[sess.clientID] {
		return errors.New(fmt.Sprintf("Client %d does not own lock at path %s", sess.clientID, path))
	}

	// Switch on lock mode.
	switch lock.mode {
	case FREE:
		// Throw an error: this means TryAcquire was not implemented correctly
		return errors.New("Lock has FREE mode: acquire not implemented correctly")
	case EXCLUSIVE:
		// Delete from lock owners
		delete(lock.owners, sess.clientID)

		// Set lock mode
		lock.mode = FREE

		// Delete lock from session locks map
		delete(sess.locks, path)

		// Return without error
		return nil
	case SHARED:
		// Delete from lock owners
		delete(lock.owners, sess.clientID)

		// Set lock mode if no more owners
		if len(lock.owners) == 1 {
			lock.mode = FREE
		}

		// Delete lock from session locks map
		delete(sess.locks, path)

		// Return without error
		return nil
	default:
		return errors.New(fmt.Sprintf("Lock at %s has undefined mode %d", path, lock.mode))
	}
}
