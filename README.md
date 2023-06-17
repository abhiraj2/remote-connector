# remote-connector
Remote File Transfer between two systems written in Golang

# Usage
### Server
* The application makes use of the ports 4242 and 4243 for control and data transfer respectively.
* Boot up the server along with one extra CLI argument which acts as the root directory. `go run server.go D:/`

### Client
* Boot up the server along with one extra CLI argument which acts as the ip address of the running server. `go run client.go 127.0.0.1`
