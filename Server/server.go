package main

import(
	"os"
	"io"
	"fmt"
	"net"
	"bufio"
	"errors"
	"encoding/json"
	"github.com/abhiraj2/remote-connector/Server/idkeeper"
	"github.com/abhiraj2/remote-connector/Server/client"
)



type Message[T any] struct{
	Status uint16
	Message T
	Msg_type string
}

func initHandshake(cl *client.Client, conn net.Conn, reader *bufio.Reader) bool{
	buf, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	var msg Message[uint16]
	json.Unmarshal([]byte(buf), &msg)
	if msg.Message != cl.Getid(){
		rep := Message[string]{500, "failed", "hs_reply"}
		reply, err:= json.Marshal(&rep)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(conn, string(append(reply, '\n')))
		return false
	} 
	rep := Message[string]{200, "ok", "hs_reply"}
	reply, err:= json.Marshal(&rep)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(conn, string(append(reply, '\n')))
	return true
}

func beginFileTransfer(cl *client.Client, path string, writer net.Conn, file *os.File) error{
	buf_size := 1024*1024
	buf := make([]byte, buf_size)
	//done := int64(0)
	transfer := true
	for transfer{
		size_read, err := file.Read(buf)
		if err!=nil{
			if err == io.EOF{
				transfer = false
			} else{
				return err
			}
		}
		writer.Write(buf[:size_read])
		if size_read < buf_size{
			transfer = false
		}
	}
	return nil
}

func handleCp(cl *client.Client, path string, serv_fs net.Listener) error{
	fmt.Println("Inside handleCp")
	conn_fs, err := serv_fs.Accept()
	if err != nil {
		panic(err)
	}
	cl_reader := bufio.NewReader(conn_fs)
	handshake:= initHandshake(cl, conn_fs, cl_reader)
	if !handshake {
		panic(errors.New("Handshake Failed"))
	}
	fmt.Println("Handshake Success")
	file, err := os.Open(path)
	if err != nil{
		panic(err)
	}
	info, _:= file.Stat()
	//cl_writer := bufio.NewWriter(conn_fs)
	msg := Message[int64]{200, info.Size(), "size"}
	msg_json, _ := json.Marshal(&msg)
	fmt.Fprintf(conn_fs, string(msg_json) + "\n")
 	err = beginFileTransfer(cl, path, conn_fs, file)
	file.Close()
	conn_fs.Close()
	return err
}

func handleCon(con net.Conn, idk *idkeeper.Idkeeper, rootpath string, serv_fs net.Listener){
	//buf := make([]byte, 1024)
	defer fmt.Println("Connection Terminated")
	r_conn:= bufio.NewReader(con)
	buf, err := r_conn.ReadString('\n')
	connect := false
	if err != nil{
		panic(err)
	} else{
		if buf == "Connect FTP\n"{
			fmt.Println("New Connection Request");
			unique_id, err := idk.AddElem()
			if err {
				panic("Error, num gen")
			}
			msg_uniqueid := &Message[uint16]{200, unique_id, rootpath}
			msg_json, error:= json.Marshal(msg_uniqueid)
			if error != nil{
				panic(error)
			}
			var cl client.Client
			cl.Setid(unique_id)
			cl.Setpwd(rootpath)
			cl.Root = rootpath
			connect = true
			fmt.Fprintf(con, string(append(msg_json, '\n')))
			var msg Message[string]
			var cmd client.Cmd
			for connect{
				cmdline, err := r_conn.ReadString('\n')
				if err != nil{
					connect = false
					continue
				}
				err = json.Unmarshal([]byte(cmdline), &msg)
				if err != nil {
					panic(err)
				}
				if msg.Msg_type == "command"{
					command_type := 1
					err := json.Unmarshal([]byte(msg.Message), &cmd)
					if err != nil{
						panic(err)
					}
					switch cmd.Argv[0]{
						case "ls":
							command_type <<= 1
						case "cd":
							command_type <<= 2
						case "cp":
							command_type <<= 3
					}
					res, err := cmd.Execute_cmd(&cl)
					//fmt.Println(res, command_type)
					if err != nil{
						msg = Message[string]{500, err.Error(), "error"}
						msg_json, _ = json.Marshal(msg)
						fmt.Fprintf(con, string(append(msg_json, '\n')))
						continue // This continue statement is really important
					}
					res_json, _ := json.Marshal(res)
					switch command_type{
						case 2:
							msg = Message[string]{200, string(res_json), "ls_reply"}
						case 4:
							msg = Message[string]{200, cl.Getpwd(), "cd_reply"}
						case 8:
							msg = Message[string]{200, res[0], "cp_reply"}
					}
					if command_type == 8 {
						go handleCp(&cl, res[0], serv_fs)
					}
					msg_json, _ = json.Marshal(msg)
					fmt.Fprintf(con, string(append(msg_json, '\n')))
				}
			}
			_ = idk.RemoveElem(unique_id)
		}
	}
}

func main() {
	//the only argument we look for is the root
	if len(os.Args) < 2{
		panic("Not Enough Args")
	}
	root := os.Args[1]
	port_control := "4242"
	port_data := "4243"
	//initiate the file tcp server
	serv_fs, err := net.Listen("tcp", ":" + port_data )
	if err != nil {
		panic(err)
	}

	// initiate the control tcp server
	serv, err := net.Listen("tcp", ":"+port_control)
	if err != nil {
		fmt.Println("Error ", err)
	} else {
		var idk idkeeper.Idkeeper
		idk.Init()
		fmt.Println("Server listening at 8000")
		for{
			con, err := serv.Accept()
			if err != nil {
				fmt.Println("Error ", err)
			} else {
				go handleCon(con, &idk, root, serv_fs)
			}
		}
	}
}
