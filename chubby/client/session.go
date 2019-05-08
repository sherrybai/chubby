package client

import (
	"net/rpc"
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
	// RPC client
	rpcClient			*rpc.Client

	// Record start time
	startTime			time.Time

	// Local timeout
	timeout				time.Duration

	// Are we in jeopardy right now?
	jeopardyFlag		bool

	// Channel for notifying if jeopardy has ended
	jeopardyChan		chan struct{}
}

const DefaultLeaseDuration time.Duration = 12 * time.Second
const JeopardyDuration time.Duration = 45 * time.Second

// Set up a Chubby session and periodically send KeepAlives to the server.
// This method should be run as a new goroutine by the client.
func InitSession(raftAddr string) (*ClientSession, error) {
	// Set up TCP connection to raftAddr.
	client, err := rpc.Dial("tcp", raftAddr)
	if err != nil {
		return nil, err
	}

	// TODO: log here

	// Initialize a session.
	sess := &ClientSession{
		rpcClient:		  client,
		startTime:		  time.Now(),
		timeout:          DefaultLeaseDuration,
		jeopardyFlag:     false,
		jeopardyChan:     nil,
	}

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
		select {
		case resp := <- keepAliveChan:
			// Process master's response
			// The master's response should contain a new, extended lease timeout.

		case <- time.After(sess.startTime.Add(sess.timeout).Sub(time.Now())):
			// Jeopardy period begins
			// If no response within local lease timeout, we have to block all RPCs
			// from the client until the jeopardy period is over.

			// Master response should give time since startTime until timeout occurs.
			// For example: extends timeout from 12s to 24s

			// Keep waiting for the response
			select {
			case resp := <- keepAliveChan:
				// Session is saved! Unblock all the requests

			case <- time.After(sess.startTime.Add(sess.timeout + JeopardyDuration).Sub(time.Now())):
				// Jeopardy period ends -- tear down the session

				break
			}
		}
	}
}

// Current plan is to implement a function for each Chubby library call.
// Each function should check jeopardyFlag to see if call should be blocked.
