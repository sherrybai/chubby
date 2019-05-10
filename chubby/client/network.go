// Assumptions the client makes about addresses of Chubby servers.

package client

// We assume that any Chubby node must have one of these addresses.
// Yes this is gross but we're doing it anyway because of time constraints
var PossibleServerAddrs = map[string]bool {
	":5379": true,
	":6379": true,
	":7379": true,
	":8379": true,
	":9379": true,
}