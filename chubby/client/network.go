// Assumptions the client makes about addresses of Chubby servers.

package client

// We assume that any Chubby node must have one of these addresses.
// Yes this is gross but we're doing it anyway because of time constraints
var PossibleServerAddrs = map[string]bool {
	"172.20.128.1:5379": true,
	"172.20.128.2:5379": true,
	"172.20.128.3:5379": true,
	"172.20.128.4:5379": true,
	"172.20.128.5:5379": true,
}