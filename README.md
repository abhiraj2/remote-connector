# remote-connector
Remote File Transfer between two systems written in Golang. Built with the usage of net library, this program enables file transfer between two systems seamless.

Current Features:
  * Support for multiple clients on the same server with usage of go routines and mutexes for synchronization.
  * Progress Bar for representing file transfer process.
  * Additional handshakes introduced for verification between the client and server pair.

# Usage
### Server
* The application makes use of the ports 4242 and 4243 for control and data transfer respectively.
* Boot up the server along with one extra CLI argument which acts as the root directory. `go run server.go D:/`
![image](https://github.com/abhiraj2/remote-connector/assets/47693983/13ba012d-3191-481c-9f51-34103a2dd270)


### Client
* Boot up the server along with one extra CLI argument which acts as the ip address of the running server. `go run client.go 127.0.0.1`
* This will give you access to the root directory of the server.


![image](https://github.com/abhiraj2/remote-connector/assets/47693983/39aaa002-6bdf-4601-ad92-fe250ad548d3)

#### Commands
* ls: works pretty much equivalent to the linux ls command `ls {path}`
![image](https://github.com/abhiraj2/remote-connector/assets/47693983/c2efaef5-121c-4045-8edc-f370f3f5001c)

* cd: equivalent to the change directory linux command. `cd path` 
* 
![image](https://github.com/abhiraj2/remote-connector/assets/47693983/088d04c8-f01b-43c8-ad0d-75f79e152240)

* cp: used for downloading and uploading. `cd server_path client_path {-u}`. Takes in an option -u flag to denote enable uploads.

![image](https://github.com/abhiraj2/remote-connector/assets/47693983/2b4ac17f-7e3d-4309-913c-40625fd4cf7d)

