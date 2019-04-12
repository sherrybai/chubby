FROM golang:1.12

WORKDIR $GOPATH/src/cos518project/
COPY . .

# Dependencies
RUN go get -d -v ./...

# Build executable
RUN go build -o bin/chubby chubby/cmd/*.go

# Expose ports
EXPOSE 5379
EXPOSE 15379    

# Run chubby exec
CMD ["bin/chubby", "-id", "id1"]