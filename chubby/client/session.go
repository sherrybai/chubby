package client

import (
	"log"
	"net/rpc"
	"os"
	"time"
	"cos518project/chubby/api"
)

type ClientSession struct {
	// Client ID
	clientID			api.ClientID

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
func InitSession(clientID api.ClientID, serverAddr string) (*ClientSession, error) {
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

	// Make RPC call.
	req := api.InitSessionRequest{ClientID: clientID}
	resp := &api.InitSessionResponse{}
	err = rpcClient.Call("Handler.InitSession", req, resp)
	if err != nil {
		return nil, err
	}

	// If the response is that the node at the address is not the leader,
	// try to send an InitSession to the leader.
	if resp.LeaderAddress != serverAddr {
		sess.logger.Printf("%s is not leader: calling InitSession to server %s", serverAddr, resp.LeaderAddress)
		return InitSession(clientID, resp.LeaderAddress)
	}

	// Call MonitorSession.
	go sess.MonitorSession()

	return sess, nil
}

func (sess *ClientSession) MonitorSession() {
	sess.logger.Printf("Monitoring session with server %s", sess.serverAddr)
	for {
		// Make new keepAlive channel.
		// This should be ok because this loop only occurs every 12 seconds to 57 seconds.
		keepAliveChan := make(chan *api.KeepAliveResponse, 1)

		// Send a KeepAlive, waiting for a response from the master.
		go func() {
			req := api.KeepAliveRequest{ClientID: sess.clientID}
			resp := &api.KeepAliveResponse{}
			err := sess.rpcClient.Call("Handler.KeepAlive", req, resp)
			if err != nil {
				sess.logger.Printf("rpc call error: %s", err.Error())
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
			sess.logger.Printf("KeepAlive response from %s received within lease timeout", sess.serverAddr)

			// Adjust new lease length.
			if (sess.leaseLength >= resp.LeaseLength) {
				sess.logger.Printf("WARNING: new lease length shorter than current lease length")
			}
			sess.leaseLength = resp.LeaseLength

		case <- time.After(durationLeaseOver):
			// Jeopardy period begins
			// If no response within local lease timeout, we have to block all RPCs
			// from the client until the jeopardy period is over.
			sess.jeopardyFlag = true
			sess.logger.Printf("session with %s in jeopardy", sess.serverAddr)

			// In a new goroutine, try to send a KeepAlive to every other server.
			// KeepAlive should check if the node is the master -> if not, ignore.
			// In KeepAlive request, eagerly send session information to server (leaseLength, locks)
			// Update session serverAddr.
			go func() {
				// TODO:
				// Jeopardy KeepAlives should allow client to eagerly send info
				// to help new leader rebuild in-mem structs
				req := api.KeepAliveRequest{
					ClientID: sess.clientID,
				}

				resp := &api.KeepAliveResponse{}

				for serverAddr := range PossibleServerAddrs {
					// Skip over current server address
					if serverAddr == sess.serverAddr {
						continue
					}

					// Try to connect to server
					rpcClient, err := rpc.Dial("tcp", serverAddr)
					if err != nil {
						sess.logger.Printf("could not dial address %s", serverAddr)
						continue
					}

					// Try to send KeepAlive to server
					sess.logger.Printf("sending KeepAlive to server %s", serverAddr)
					err = rpcClient.Call("Handler.KeepAlive", req, resp)
					if err == nil {
						// Successfully contacted new leader!
						sess.logger.Printf("received KeepAlive resp from server %s", serverAddr)

						// Update session details
						sess.serverAddr = serverAddr
						sess.rpcClient = rpcClient

						// Send response onto channel
						keepAliveChan <- resp

						break  // Avoid closing new rpc client
					} else {
						sess.logger.Printf("KeepAlive error from server at %s: %s", serverAddr, err.Error())
					}

					rpcClient.Close()
				}
			}()

			// Wait for responses.
			select {
			case resp := <- keepAliveChan:
				// Session is saved! Unblock all the requests
				sess.jeopardyFlag = false
				sess.jeopardyChan <- sess.jeopardyFlag
				sess.logger.Printf("session with %s safe", sess.serverAddr)

				// Process master's response
				// Adjust new lease length.
				if (sess.leaseLength >= resp.LeaseLength) {
					sess.logger.Printf("WARNING: new lease length shorter than current lease length")
				}
				sess.leaseLength = resp.LeaseLength

			case <- time.After(durationJeopardyOver):
				// Jeopardy period ends -- tear down the session
				sess.expired = true
				close(keepAliveChan)
				err := sess.rpcClient.Close()
				if err != nil {
					sess.logger.Printf("rpc close error: %s", err.Error())
				}
				sess.logger.Printf("session with %s expired", sess.serverAddr)
				return
			}
		}

		// Close the channel.
		close(keepAliveChan)
	}
}

// Current plan is to implement a function for each Chubby library call.
// Each function should check jeopardyFlag to see if call should be blocked.
