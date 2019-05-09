package client

import (
	"log"
	"net/rpc"
	"os"
	"time"
)

type ClientID string

type InitSessionRequest struct {
	clientID ClientID
}

type InitSessionResponse struct {
	leaderAddress string
	isLeader bool
}

type KeepAliveRequest struct {
	clientID ClientID
}

type KeepAliveResponse struct {
	leaseLength time.Duration
}

type ClientSession struct {
	// Client ID
	clientID			ClientID

	// Server address
	serverAddr			string

	// RPC client
	rpcClient			*rpc.Client

	// Record start time
	startTime			time.Time

	// Local lease length
	leaseLength			time.Duration

	// Are we in jeopardy right now?
	jeopardyFlag		bool

	// Channel for notifying if jeopardy has ended
	jeopardyChan		chan bool

	// Did this session expire?
	expired				bool

	// Logger
	logger				*log.Logger
}

const DefaultLeaseDuration time.Duration = 12 * time.Second
const JeopardyDuration time.Duration = 45 * time.Second

// Set up a Chubby session and periodically send KeepAlives to the server.
// This method should be run as a new goroutine by the client.
func InitSession(clientID ClientID, serverAddr string) (*ClientSession, error) {
	// Initialize a session.
	sess := &ClientSession{
		clientID:     clientID,
		serverAddr:   serverAddr,
		startTime:    time.Now(),
		leaseLength:  DefaultLeaseDuration,
		jeopardyFlag: false,
		jeopardyChan: nil,
		expired:      false,
		logger:       log.New(os.Stderr, "[client] ", log.LstdFlags),
	}

	sess.logger.Printf("session with %s initialized at client", serverAddr)

	// Call InitSession at server.
	// Set up TCP connection to serverAddr.
	rpcClient, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	sess.rpcClient = rpcClient
	defer rpcClient.Close()

	// Make RPC call.
	req := InitSessionRequest{clientID: clientID}
	resp := &InitSessionResponse{}
	err = rpcClient.Call("Handler.InitSession", req, resp)
	if err != nil {
		return nil, err
	}

	// If the response is that the node at the address is not the leader,
	// try to send an InitSession to the leader.
	if !resp.isLeader {
		sess.logger.Printf("%s is not leader: calling InitSession to server %s", serverAddr, resp.leaderAddress)
		return InitSession(clientID, resp.leaderAddress)
	}

	// Call MonitorSession.
	go sess.MonitorSession()

	return sess, nil
}

func (sess *ClientSession) MonitorSession() {
	for {
		// Make new keepAlive channel.
		// This should be ok because this loop only occurs every 12 seconds to 57 seconds.
		keepAliveChan := make(chan *KeepAliveResponse, 1)

		// Send a KeepAlive, waiting for a response from the master.
		go func() {
			req := KeepAliveRequest{clientID: sess.clientID}
			resp := &KeepAliveResponse{}
			err := sess.rpcClient.Call("Handler.KeepAlive", req, resp)
			if err != nil {
				sess.logger.Printf("rpc dial error: %s", err.Error())
				return  // do not push anything onto channel -- session will time out
			}

			keepAliveChan <- resp
		}()

		// Set up timeout
		durationLeaseOver := time.Until(sess.startTime.Add(sess.leaseLength))
		durationJeopardyOver := time.Until(sess.startTime.Add(sess.leaseLength + JeopardyDuration))

		select {
		case resp := <- keepAliveChan:
			// Process master's response
			// The master's response should contain a new, extended lease timeout.
			sess.logger.Printf("session with %s received within lease timeout", sess.serverAddr)

			// Adjust new lease length.
			if (sess.leaseLength >= resp.leaseLength) {
				sess.logger.Printf("WARNING: new lease length shorter than current lease length")
			}
			sess.leaseLength = resp.leaseLength

		case <- time.After(durationLeaseOver):
			// Jeopardy period begins
			// If no response within local lease timeout, we have to block all RPCs
			// from the client until the jeopardy period is over.
			sess.jeopardyFlag = true
			sess.logger.Printf("session with %s in jeopardy", sess.serverAddr)

			// TODO:
			// In a new goroutine, try to send a KeepAlive to every other server.
			// KeepAlive should check if the node is the master -> if not, ignore.
			// In KeepAlive request, eagerly send session information to server (leaseLength, locks)
			// Update session serverAddr.

			select {
			case resp := <- keepAliveChan:
				// Session is saved! Unblock all the requests
				sess.jeopardyFlag = false
				sess.jeopardyChan <- sess.jeopardyFlag
				sess.logger.Printf("session with %s safe", sess.serverAddr)

				// Process master's response
				// Adjust new lease length.
				if (sess.leaseLength >= resp.leaseLength) {
					sess.logger.Printf("WARNING: new lease length shorter than current lease length")
				}
				sess.leaseLength = resp.leaseLength

			case <- time.After(durationJeopardyOver):
				// Jeopardy period ends -- tear down the session
				sess.expired = true
				close(keepAliveChan)
				sess.logger.Printf("session with %s expired", sess.serverAddr)
				break
			}
		}

		// Close the channel.
		close(keepAliveChan)
	}
}

// Current plan is to implement a function for each Chubby library call.
// Each function should check jeopardyFlag to see if call should be blocked.
