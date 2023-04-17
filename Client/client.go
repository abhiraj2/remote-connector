package main

import(
	"os"
	"io"
	"fmt"
	"net"
	"bufio"
	"errors"
	"encoding/json"
	"github.com/abhiraj2/remote-connector/Client/cmdline"
	"github.com/schollz/progressbar/v3"
)

// Message provides a standard layout through which communication between the client and server takes place
type Message[T any] struct{
	Status uint16
	Message T
	Msg_type string
}

type Client struct{
	Client_id uint16
	Pwd string
	Root string
}

func print_ls(lis []string){
	for _, name := range lis{
		fmt.Println(name)
	}
	fmt.Println("-----------------------------END---------------------------------")
}

func initHandshake(id uint16, conn net.Conn, reader *bufio.Reader) bool {
	msg := Message[uint16]{200, id, "handshake"}
	msg_json, err := json.Marshal(&msg)
	fmt.Fprintf(conn, string(append(msg_json,  '\n')))
	buf, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	var reply Message[string]
	json.Unmarshal([]byte(buf), &reply)
	if reply.Status != 200{
		return false
	}
	return true
}

func BeginFileTransferToSever(cl *Client, cmd *cmdline.Cmd, writer net.Conn, file *os.File, file_size int64) error{
	buf_size := 1024*1024
	buf := make([]byte, buf_size)
	transfer := true
	bar := progressbar.Default(file_size)
	for transfer{
		size_read, err := file.Read(buf)
		if err!=nil{
			if err == io.EOF{
				transfer = false
			} else{
				return err
			}
		}
		bar.Add(size_read)
		writer.Write(buf[:size_read])
		if size_read < buf_size{
			transfer = false
		}
	}
	return nil
}

func BeginFileTransferFromSever(cl *Client, cmd *cmdline.Cmd, reader net.Conn, file_size int64) error{
	buf_size := 1024
	buf := make([]byte, buf_size)
	transfer := true
	new_file, err := os.Create(cmd.Argv[2])
	if err != nil {
		panic(err)
	}
	done := int64(0)
	bar := progressbar.Default(file_size);
	for transfer {
		size_read, err := reader.Read(buf)
		if err != nil{
			return err
		}
		done += int64(size_read)
		bar.Add(size_read)

		new_file.Write(buf[:size_read])
		if done >= file_size{
			fmt.Println("File Transfer Done")
			transfer = false
		}
	}
	new_file.Close()
	return nil
}

func cp_handler(cl *Client, path string, cmd *cmdline.Cmd, ip string, port string, upload bool) error{
	fmt.Println("inside client cp_handler")
	conn_fs, err := net.Dial("tcp", ip+":"+port)
	if err != nil{
		panic(err)
	}
	serv_reader := bufio.NewReader(conn_fs)
	handshake := initHandshake(cl.Client_id, conn_fs, serv_reader)
	if !handshake{
		panic(errors.New("Handshake Failed"))
	}
	fmt.Println("Handshake Success")
	if upload {
		file, err := os.Open(path)
		defer file.Close()
		if err != nil{
			panic(err)
		}
		//Get file info for getting the size and send a corresponding message to the client
		info, _:= file.Stat()
		msg := Message[int64]{200, info.Size(), "size"}
		msg_json, _ := json.Marshal(&msg)
		fmt.Fprintf(conn_fs, string(msg_json) + "\n")
		err = BeginFileTransferToSever(cl, cmd, conn_fs, file, info.Size())
	} else{
		msg := Message[int64]{}
		msg_json, _ := serv_reader.ReadString('\n')
		json.Unmarshal([]byte(msg_json), &msg)
		file_size := msg.Message
		fmt.Println(file_size)
		err = BeginFileTransferFromSever(cl, cmd, conn_fs, file_size)
	}
	
	return nil
}


//ftp_start starts the FTP session with an interactive terminal
//returns nothing
func ftp_start(conn net.Conn, uniqueclientid uint16, rootpath string, ip string, port string){
	fmt.Println("Session Started")
	terminal := true
	server_reader := bufio.NewReader(conn)
	var command cmdline.Cmd
	var err error
	var msg Message[string]
	inp := bufio.NewReader(os.Stdin)
	cl := Client{uniqueclientid, rootpath, rootpath}
	for terminal{
		fmt.Printf("%v\t%v>: ", uniqueclientid, cl.Pwd)
		command.Command, err = inp.ReadString('\n')
		if err != nil{
			panic(err)
		}
		command.Parsecmdline()
		valid := command.CheckValid()
		if valid{
			if command.Argv[0] == "quit"{
				terminal = false
			} else{
				command_json, erro:= json.Marshal(command)
				if erro != nil{
					panic(erro)
				}
				msg = Message[string]{200, string(command_json), "command"}
				msg_json, erro := json.Marshal(msg)
				fmt.Fprintf(conn, string(append(msg_json, '\n')))
				res, erro:= server_reader.ReadString('\n')
				if erro != nil {
					panic(erro)
				}
				var reply_msg Message[string]
				json.Unmarshal([]byte(res), &reply_msg)
				if reply_msg.Status != 200 {
					fmt.Println(reply_msg.Message)
					continue
				}
				switch reply_msg.Msg_type{
					case "ls_reply":
						var reply []string
						json.Unmarshal([]byte(reply_msg.Message),&reply)
						print_ls(reply)
					case "cd_reply":
						reply := reply_msg.Message
						cl.Pwd = reply
					case "cp_reply":
						_ = reply_msg.Message
						u:= false
						if len(command.Argv) > 3 && command.Argv[3] == "-u"{
							u = true
						}
						cp_handler(&cl, command.Argv[2], &command, ip, port, u)
				}
			}
		} else{
			fmt.Println("Please enter a valid command")
		}
	}
}

func main(){
	if len(os.Args) < 2{
		panic("Not Enough Args")
	}
	ip := os.Args[1]

	port_control := "4242"
	port_data := "4243"

	conn, err := net.Dial("tcp", ip+":"+port_control)
	if err != nil{
		fmt.Println(err)
	} else{
		var uniqueclientid uint16
		fmt.Fprintf(conn, "Connect FTP\n");
		server_reader := bufio.NewReader(conn)
		msg_json, err := server_reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		msg := Message[uint16]{}
		json.Unmarshal([]byte(msg_json[:len(msg_json)]), &msg)
		if msg.Status == 200{
			fmt.Println("Connection Established, unique client id: ", msg.Message);
			uniqueclientid = msg.Message
			rootpath := msg.Msg_type // this is a scam, I had to pull it off sorry 😢😭😭
			ftp_start(conn, uniqueclientid, rootpath, ip, port_data)
		}
	}
}