/*
Client package includes a lot of functions that make working with the client really easy.
The package contain a lot of functionality with the client side of things
*/

package client

import(
	"os"
	"fmt"
	"errors"
	"io/ioutil"
)

type Cmd struct{
	Command string
	Argv []string
}

//Execute_cmd recognises the command and calls the aoppropriate string
//returns a list of strings along with error if any
func (cmd *Cmd) Execute_cmd(cl *Client) ([]string, error){
	fmt.Println(cmd.Command)
	switch cmd.Argv[0] {
		case "ls":
			res, err := cl.Ls(cmd)
			if err != nil{
				return nil, err
			} else{
				return res, nil
			}
		case "cd": 
			_, err := cl.Cd(cmd)
			if err != nil {
				return nil, err
			}
			return nil, nil
		case "cp":
			res, err := cl.Cp(cmd)
			if err != nil {
				return nil, err
			}
			return res, nil
	}
	return nil, nil
}

type Client struct{
	client_id uint16
	pwd string
	Root string
}

//Getters and Setters
func (cl *Client) Setid(id uint16){
	cl.client_id = id
}

func (cl *Client) Getid() uint16{
	return cl.client_id
}

func (cl *Client) Setpwd(pwd string){
	cl.pwd = pwd
}

func (cl *Client) Getpwd() string{
	return cl.pwd
}

func check_path_valid(path string) (bool, bool){
	stat, err:=os.Stat(path)
	if err != nil{
		if os.IsNotExist(err){
			return false, false
		} else{
			panic(err)
		}
	}
	//fmt.Println(stat)
	return true, stat.IsDir()
}

//Parse_and_return_new_path parses the string passed to it as a path
//takes in path along with root and present working directory
//returns the parsed string
func Parse_and_return_new_path(path string, root string, pwd string) string{
	// absolute or relative
	// ends in slash or not res will always end in /
	var res string
	vec := 0
	
	//absolute?
	if path[0] == '/'{
		vec = vec | 1
	}
	//ends with a /
	if path[len(path)-1] == '/'{
		vec = vec | 2
	}
	
	if (vec & 1) != 0 {
		//if absolute append it to the root
		res = root + path[1:]
	} else{
		//if relative, work on all the occurrences of ./ and ../ 
		flag := 0
		for idx,ch := range path{
			if ch == '.' && (path[idx+1] == '/' || path[idx+1] == '.'){
				flag++
			} else if flag == 2 && ch == '/'{
				flag = 0
				temp := ""
				ins := false
				for i := len(pwd)-2; i>=0; i--{
					if(pwd[i] == '/'){
						ins = true
					}
					if ins{
						temp = string(pwd[i]) + temp
					}					
				} 
				pwd = temp
			} else if flag == 1 && ch == '/'{
				continue
			} else{
				flag =0
				pwd = pwd + string(ch)
			}
		}
		res = pwd
	}
	_, isDir := check_path_valid(res)
	if (vec & 2) == 0 && isDir{
		res = res + "/"
	}
	return res
}


//Ls is somewhat like general Unix ls command
//it needs a second argument describing the path relative to pwd or absolute to root
//No support for back dirs.
//returns a list of filenames on success
func (cl *Client) Ls(cmd *Cmd) ([]string, error){
	path := "./"
	if len(cmd.Argv) > 1{
		path = cmd.Argv[1]	
	}
	path = Parse_and_return_new_path(path, cl.Root, cl.Getpwd())
	valid, _ := check_path_valid(path)
	if !valid {
		return nil, errors.New("Invalid Pathname")
	}
	var filenames []string
	fileinfo, err := ioutil.ReadDir(path)
	if err != nil{
		return nil, err
	}
	for _, file := range fileinfo{
		if file.IsDir() {
			filenames = append(filenames, file.Name()+ "--directory")
		} else{
			filenames = append(filenames, file.Name())
		}
		
	}
	//fmt.Println(filenames)
	return filenames, nil

}


//Cd tries to simulate the Unix cd program but works a little different
//currently ../ is broken
//Absolute paths start with a blank instead of /
//Relative paths as usual start with ./ but only go forward
//On success returns nil nil
func (cl *Client) Cd(cmd *Cmd)([]string, error){
	if len(cmd.Argv) < 2{
		return nil, errors.New("Not enough arguments")
	}
	path := cmd.Argv[1]
	path = Parse_and_return_new_path(path, cl.Root, cl.Getpwd())
	valid, _:= check_path_valid(path)
	if !valid {
		return nil, errors.New("Invalid Pathname")
	}
	//path = strings.Replace(path, "./", "", len(path))
	//fmt.Println(path)
	cl.Setpwd(path)
	return nil, nil
}


func check_valid_file(path string) bool{
	stat, err:=os.Stat(path)
	if err != nil{
		if os.IsNotExist(err){
			return false
		} else{
			panic(err)
		}
	}
	if stat.IsDir(){
		return false
	}
	//fmt.Println(stat)
	return true
}


//Cp performs a check on the command and if it is valid or not 
//returns file path if it is valid
func (cl *Client) Cp(cmd *Cmd) ([]string, error){
	var res []string
	//cp needs to have two args
	if len(cmd.Argv) < 3{
		return nil, errors.New("Not enough arguments")
	}
	
	path := cmd.Argv[1]
	path = Parse_and_return_new_path(path, cl.Root, cl.Getpwd())
	valid := true
	if len(cmd.Argv) <= 3{
		valid = check_valid_file(path)
	}
	if !valid {
		return nil, errors.New("Invalid Pathname or the path is not a file")
	}
	return append(res, path), nil
}