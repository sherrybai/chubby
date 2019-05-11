all: chubby simple_client client_fast_reqs

chubby:
	go build -o chubby cmd/main.go

simple_client:
	go build -o simple_client cmd/simple_client.go

client_fast_reqs:
	go build -o client_fast_reqs cmd/client_fast_reqs.go

clean:
	rm chubby simple_client client_fast_reqs
