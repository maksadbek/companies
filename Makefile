all: server client

server:
	cd src && go build -o ../bin/server ./cmd/server

client:
	cd src && go build -o ../bin/client ./cmd/client
