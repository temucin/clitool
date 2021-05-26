package CmdRegistry

import (
	"flag"
	"os"
)

type Cmd struct {
	Name    string
	RunCmd  func()
	FlagSet flag.FlagSet
}

var Cmds = []Cmd{}
var FlagSets = []flag.FlagSet{}

func RegisterCmd(cmd Cmd) {
	Cmds = append(Cmds, cmd)
}

func RegisterFlagSet(fs flag.FlagSet) {
	FlagSets = append(FlagSets, fs)
}

func CmdArgs() []string {
	return os.Args[2:]
}
