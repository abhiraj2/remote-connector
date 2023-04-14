package cmdline

import (
	"strings"
	"fmt"
)

var Help = "Allowed Commands:\n\t*ls\n\t*cp\n\t*cd\n\t*quit\n"

// Cmd provides standard layout for the command that are going to be given to the terminal
type Cmd struct{
	Command string
	Argv []string
}

// Parsecmdline is will work on the Cmd struct
// fills in the Argv member and returns nothing
func (cmd *Cmd) Parsecmdline(){
	cmd.Argv = strings.Fields(cmd.Command)
}

func(cmd *Cmd) CheckValid() bool{
	switch cmd.Argv[0] {
		case "ls":
			return true
		case "cd": 
			return true
		case "quit":
			return true
		case "cp":
			return true
		default:
			fmt.Print(Help)
			return false
	}
}
