// Defines Chubby session metadata, as well as operations on locks that can
// be performed as part of a Chubby session.

package server

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"cos518project/chubby/api"
)

const DefaultLeaseExt = 12 * time.Second

// Session contains metadata for one Chubby session.
// For simplicity, we say that each client can only init one session with
// the Chubby servers.
type Session struct {
	// Client to which this Session corresponds.
	clientID 		api.ClientID

	// Start time
	startTime		time.Time

	// Length of the Lease
	leaseLength     time.Duration

	//TTL Lock
	ttlLock 		sync.Mutex

	// Channel used to block KeepAlive
	ttlChannel  	chan struct{}

    // A data structure describing which locks the client holds.
    // Maps lock filepath -> Lock struct.
    locks           map[api.FilePath]*Lock

	// Did we terminate this session?
	terminated		bool

	// Terminated channel
	terminatedChan	chan struct{}
}

// Lock describes information about a particular Chubby lock.
type Lock struct {
	path			api.FilePath  // The path to this lock in the store.
	mode			api.LockMode  // api.SHARED or exclusive lock?
	owners			map[api.ClientID]bool  // Who is holding the lock?
}

/* Create Session struct. */
func CreateSession(clientID api.ClientID) (*Session, error) {
	sess, ok := app.sessions[clientID]

	if ok && !sess.terminated {
		return nil, errors.New(fmt.Sprintf("The client already has a session established with the master"))
	}



	app.logger.Printf("Creating session with client %s", clientID)

	// Create new session struct.
	sess = &Session{
        clientID:    	clientID,
        startTime:   	time.Now(),
        leaseLength: 	DefaultLeaseExt,
        ttlChannel:  	make(chan struct{}, 2),
        locks:       	make(map[api.FilePath]*Lock),
        terminated:	 	false,
        terminatedChan: make(chan struct{}, 2),
    }

	// Add the session to the sessions map.
	app.sessions[clientID] = sess

	// In a separate goroutine, periodically check if the lease is over
    go sess.MonitorSession()

    return sess, nil
}

func (sess *Session) MonitorSession() {
	app.logger.Printf("Monitoring session with client %s", sess.clientID)

	// At each second, check time until the lease is over.
	ticker := time.Tick(time.Second)
	for range ticker {
		timeLeaseOver := sess.startTime.Add(sess.leaseLength)

		var durationLeaseOver time.Duration = 0
		if timeLeaseOver.After(time.Now()) {
			durationLeaseOver = time.Until(timeLeaseOver)
		}

		if durationLeaseOver == 0 {
			// Lease expired: terminate the session
			app.logger.Printf("Lease with client %s expired: terminating session", sess.clientID)
			sess.TerminateSession()
			return
		}

		if durationLeaseOver <= (1 * time.Second) {
			// Trigger KeepAlive response 1 second before timeout
			sess.ttlChannel <- struct{}{}
		}
	}
}

// Terminate the session.
func (sess *Session) TerminateSession() {
	// We can't justs delete the session from the app session map because
	// We cannot delete the session from the app session map because
	// Chubby could have experienced a failover event.
	sess.terminated = true
	close(sess.terminatedChan)

	// Release all the locks in the session.
	for filePath := range sess.locks {
		err := sess.ReleaseLock(filePath)
		if err != nil {
			app.logger.Printf(
				"error when client %s releasing lock at %s: %s",
				sess.clientID,
				filePath,
				err.Error())
		}
	}

	app.logger.Printf("terminated session with client %s", sess.clientID)
}

// Extend Lease after receiving keepalive messages
func (sess *Session) KeepAlive(clientID api.ClientID) (time.Duration) {
	// Block until shortly before lease expires
	select {
	case <- sess.terminatedChan:
		// Return early response saying that session should end.
		return sess.leaseLength

	case <- sess.ttlChannel:
		// Extend lease by 12 seconds
		sess.leaseLength = sess.leaseLength + DefaultLeaseExt

		app.logger.Printf(
			"session with client %s extended: lease length %s",
			sess.clientID,
			sess.leaseLength.String())

		// Return new lease length.
		return sess.leaseLength
	}
}

// Create the lock if it does not exist.
func (sess *Session) OpenLock(path api.FilePath) error {
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
			mode: api.FREE,
			owners: make(map[api.ClientID]bool),
		}
	}

	return nil
}

// Delete the lock. Lock must be held in exclusive mode before calling DeleteLock.
func (sess *Session) DeleteLock(path api.FilePath) error {
	// If we are not holding the lock, we cannot delete it.
	lock, exists := sess.locks[path]
	if !exists {
		return errors.New(fmt.Sprintf("Client does not hold the lock at path %s", path))
	}

	// Check if we are holding the lock in exclusive mode
	if lock.mode != api.EXCLUSIVE {
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
func (sess *Session) TryAcquireLock (path api.FilePath, mode api.LockMode) (bool, error) {
	// Validate mode of the lock.
	if mode != api.EXCLUSIVE && mode != api.SHARED {
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
		app.logger.Printf("Lock Doesn't Exist with Client ID", sess.clientID)

		app.locks[path] = &Lock{
			path: path,
			mode: api.FREE,
			owners: make(map[api.ClientID]bool),
		}
		lock = app.locks[path]
	}

	// Check the mode of the lock
	switch lock.mode {
	case api.EXCLUSIVE:
		// Should fail: someone probably already owns the lock
		if len(lock.owners) == 0 {
			// Throw an error if there are no owners but lock.mode is api.EXCLUSIVE:
			// this means ReleaseLock was not implemented correctly
			return false, errors.New("Lock has EXCLUSIVE mode despite having no owners")
		} else if len(lock.owners) > 1 {
			return false, errors.New("Lock has EXCLUSIVE mode but has multiple owners")
		} else {
			// Fail with no error
			app.logger.Printf("Failed to acquire lock %s: already held in EXCLUSIVE mode", path)
			return false, nil
		}
	case api.SHARED:
		// If our mode is api.SHARED, then succeed; else fail
		if mode == api.EXCLUSIVE {
			app.logger.Printf("Failed to acquire lock %s in EXCLUSIVE mode: already held in SHARED mode", path)
			return false, nil
		} else {  // mode == api.SHARED
			// Update lock owners
			lock.owners[sess.clientID] = true

			// Add lock to session lock struct
			sess.locks[path] = lock

			// Return success
			//app.logger.Printf("Lock %s acquired successfully with mode SHARED", path)
			return true, nil
		}
	case api.FREE:
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

		// Update Lock Mode in the global Map
		app.locks[path] = lock

		// Return success
		//if mode == api.SHARED {
		//	app.logger.Printf("Lock %s acquired successfully with mode SHARED", path)
		//} else {
		//	app.logger.Printf("Lock %s acquired successfully with mode EXCLUSIVE", path)
		//}
		return true, nil
	default:
		return false, errors.New(fmt.Sprintf("Lock at %s has undefined mode %d", path, lock.mode))
	}
}

// Release the lock.
func (sess *Session) ReleaseLock (path api.FilePath) (error) {
	// Check if lock exists in persistent store
	_, err := app.store.Get(string(path))

	if err != nil {
		return errors.New(fmt.Sprintf("Client with id %s: Lock at %s does not exist in persistent store", path, sess.clientID))
	}

	// Grab lock struct from session locks map.
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
	case api.FREE:
		// Throw an error: this means TryAcquire was not implemented correctly
		return errors.New(fmt.Sprint("Lock at %s has FREE mode: acquire not implemented correctly Client ID %s", path, sess.clientID))
	case api.EXCLUSIVE:
		// Delete from lock owners
		delete(lock.owners, sess.clientID)

		// Set lock mode
		lock.mode = api.FREE

		// Delete lock from session locks map
		delete(sess.locks, path)
		app.locks[path] = lock

		// Return without error
		return nil
	case api.SHARED:
		// Delete from lock owners
		delete(lock.owners, sess.clientID)

		// Set lock mode if no more owners
		if len(lock.owners) == 1 {
			lock.mode = api.FREE
		}

		// Delete lock from session locks map
		delete(sess.locks, path)
		app.locks[path] = lock

		// Return without error
		return nil
	default:
		return errors.New(fmt.Sprintf("Lock at %s has undefined mode %d", path, lock.mode))
	}
}
