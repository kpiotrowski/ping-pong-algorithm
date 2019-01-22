build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/ping-pong ping-pong/main.go

