# ping-pong-algorithm
Simple implementation of the Mutual exclusion within ring topology(Ping-Pong) algorithm. It uses ZeroMQ for the network communication.

## Build
Make sure that ZeroMQ is installed in you system (version >= 4.0.1).
```
dep ensure
make build
```

## Test

### Run ping-pong node

```
./bin/ping-pong (flags) [node_addr:port] [next_node_addr:port]
```
example:
```
./bin/ping-pong 127.0.0.1:8880 127.0.0.1:8881
./bin/ping-pong 127.0.0.1:8881 127.0.0.1:8882
./bin/ping-pong 127.0.0.1:8882 127.0.0.1:8880
```

You can also add additional flags to change algorithm behavior:

- 