package client

import (
	"log"
	"net/rpc"
	"os"
	"time"
)

type ClientRequest struct {
	clientID 	string
	params 		[]byte
}

type ClientResponse struct {
	response	[]byte
}

type ClientSession struct {
	// Server address
	serverAddr			string

	// RPC client
	rpcClient			*rpc.Client

	// Record start time
	startTime			time.Time

	// Local lease length
	leaseLength				time.Duration

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
func InitSession(serverAddr string) (*ClientSession, error) {
	// Set up TCP connection to serverAddr.
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	// Initialize a session.
	sess := &ClientSession{
		serverAddr:		  serverAddr,
		rpcClient:		  client,
		startTime:		  time.Now(),
		leaseLength:      DefaultLeaseDuration,
		jeopardyFlag:     false,
		jeopardyChan:     nil,
		expired:		  false,
		logger:			  log.New(os.Stderr, "[client] ", log.LstdFlags),
	}

	sess.logger.Printf("session with %s initialized", serverAddr)

	// TODO: fill out req/resp
	req := ClientRequest{}
	resp := ClientResponse{}
	err = client.Call("Handler.InitSession", req, resp)

	// If the response is that the node at the address is not the leader,
	// try to send an InitSession to the leader.

	// Otherwise, throw an error.

	// Call MonitorSession.
	go sess.MonitorSession()

	return sess, nil
}

func (sess *ClientSession) MonitorSession() {
	var err error

	for {
		// Make new keepAlive channel.
		// This should be ok because this loop only occurs every 12 seconds to 57 seconds.
		keepAliveChan := make(chan ClientResponse, 1)

		// Send a KeepAlive, waiting for a response from the master.
		go func() {
			// TODO: fill out req/resp
			req := ClientRequest{}
			resp := ClientResponse{}
			err = sess.rpcClient.Call("Handler.KeepAlive", req, resp)
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


		case <- time.After(durationLeaseOver):
			// Jeopardy period begins
			// If no response within local lease timeout, we have to block all RPCs
			// from the client until the jeopardy period is over.
			sess.jeopardyFlag = true
			sess.logger.Printf("session with %s in jeopardy", sess.serverAddr)

			// Keep waiting for the response
			select {
			case resp := <- keepAliveChan:
				// Session is saved! Unblock all the requests
				sess.jeopardyFlag = false
				sess.jeopardyChan <- sess.jeopardyFlag
				sess.logger.Printf("session with %s safe", sess.serverAddr)

				// Process master's response


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
