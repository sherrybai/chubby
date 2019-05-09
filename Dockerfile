FROM golang:1.12

WORKDIR $GOPATH/src/cos518project/
COPY . .

# Dependencies
RUN go get -d -v ./...

# Build executable
RUN go build -o bin/simple_client cmd/simple_client.go

# Run simple_client exec
CMD ["bin/simple_client"]