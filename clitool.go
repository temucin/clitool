package main

import (
	_ "clitool/cmd/assume"
	_ "clitool/cmd/elastic"
	_ "clitool/cmd/kssh"
	"clitool/utils/CmdRegistry"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/chzyer/readline"
)

var interactive bool
var mainFlagSet flag.FlagSet
var cmd string
var args []string

func init() {
	mainFlagSet = *flag.NewFlagSet("main", flag.ContinueOnError)
	mainFlagSet.BoolVar(&interactive, "i", false, "Specifies whether CanopyCLI should be run in interactive mode or not.")
}

func main() {

	mainFlagSet.Parse(os.Args[1:])

	if interactive {

		fmt.Println("--- INTERACTIVE MODE ---") //TODO(Print something more awesome and lulz worthy)

		rl, err := readline.NewEx(&readline.Config{
			Prompt:      "\033[34mclitool>\033[0m ",
			HistoryFile: "/tmp/readline.tmp",
			EOFPrompt:   "exit",
		})

		if err != nil {
			panic(err)
		}
		defer rl.Close()

		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}
			lineSplit := strings.Fields(line)
			cmd = lineSplit[0]   //get command from read line
			args = lineSplit[1:] //get command arguments
			processCmd(cmd, args)
		}

	} else {
		cmd = os.Args[1]
		args = os.Args[2:]
		processCmd(cmd, args)
	}

}

func processCmd(cmd string, args []string) {
	//Watches for the recover function to bubble errors up to the user and print a stack trace.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error running command:", r)
			fmt.Printf("%s\n", debug.Stack())
		}
	}()

	switch cmd {
	case "help":
		processHelp()
	case "exit":
		processExit()
	default:
		for _, c := range CmdRegistry.Cmds {
			if cmd == c.Name {
				c.RunCmd()
				return
			}
		}
		fmt.Println("Command not found! Run help to see all commands and flags.")
	}
}

func processHelp() {
	for _, fs := range CmdRegistry.FlagSets {
		fmt.Println("Command: " + fs.Name())
		fmt.Print("Usage: ")
		fs.Usage()
		fmt.Print("\nFlags and Arguments\n")
		fs.PrintDefaults()
	}
	fmt.Println("")
	fmt.Println("Command: ksftp\nUsage: Same usage as kssh but executes MSFTP instead of MSSH")
}

func processExit() {
	fmt.Println("Goodbye!")
	os.Exit(0)
}
