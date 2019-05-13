all: chubby simple_client client_fast_reqs test_client over_load_client leader_election_client1 acquire_lock_client

chubby:
	go build -o chubby cmd/main.go

simple_client:
	go build -o simple_client cmd/simple_client.go

client_fast_reqs:
	go build -o client_fast_reqs cmd/client_fast_reqs.go

over_load_client:
	go build -o over_load_client cmd/overload_leader_client.go

clean:
	rm chubby simple_client client_fast_reqs test_client over_load_client leader_election_client1 acquire_lock_client

test_client:
	go build -o test_client cmd/testLock_client.go

leader_election_client1:
	go build -o leader_election_client1 cmd/leader_election.go

acquire_lock_client:
    go build -o acquire_lock_client cmd/acquire_Lock_Client.go
