# CLI Tool

This tool is based on an original command line interface I wrote for interfacing with AWS utilities within a proprietary project. It has been recreated as a shell so I may reuse it again in other projects.

The tool's primary purpose was to allow for an ease of switching between AWS roles. The codebase grew to support extensions to interface with other AWS services. Two samples have been included based on previous ideas, an SSH utility and an ElasticSearch utility.


# Running the Tool
The CLI is runnable as an executable binary. It can also be run from soure code using the "go run" command. Note that the binary must be built for the specific OS in which it is intended to be run against.

To use the above command you need to init the cli as a Go module named  "clitool". Use the following command to do so:

`go mod init clitool`


## Running Commands

Simply write the command and its flag arguments after the executable to execute the command

`./clitool kssh -t green -a Frontend -e Dev`
`go run clitool kssh -t green -a Frontend -e Dev`
`clitool> kssh -t green -a Frontend -e Dev`

Pass in "help" as an argument to see the list of commands.

`go run clitool.go help`

To run the tool in interactive mode, pass in the "-i" flag. 

`clitool -i`

## Developing a Command

Each utility is defined as a cmd inside the "cmd" directory. A command can be any Go code and the intention of the command is left up to the implementer and use case.

The CLI has been written using a Command registry pattern so that the CLI may be easily extended. To create a custom command:

1. Create a directory for your command in the cmd directory. 
2. Implement the Cmd struct from CmdRegistry where name is the desired Name of your command and RunCmd is the main entry point to your logic.
3. Creating a FlagSet for your command is not required, but highly recommended to work well with the CLI and be involved during the help function. The Usage(), Name(), and PrintDefaults() functions will be used by the main help command.
4. In the init function of your command, call the RegisterCmd and RegisterFlagSet from the CmdRegistry. This will be used by the CLIs main code to identify your command when called and run its main code as you've specified it.
5. Finally, import your command within clitool.go into the unused variable. When the CLI is run, it will call the init function of your command, thus registering it with the CmdRegistry, and allow the CLI to execute its functionality as described above. 

Note: Go Plugins could have more easily been used to replicate the above behavior but at the time of this writing, plugins are not supported on Windows. 

#### THINGS TODO
1. ~~Create proper help printout from command.~~ 
1. ~~Improve help printout behavior to iterate over every commands in the commandss directory and use the flagset within each commands to print help information.~~
1. Create a proper test functionality and test cases.
2. ~~Properly process '-h' and '-help' flags for commandss.~~
3. ~~Figure out how to use "usage" command for each commands.~~
3. ~~Improve use of 'Usage' for each commands.~~
4. ~~Work on other commands.~~
5. ~~Update readme to include examples.~~
5. ~~Update readme to include information on writing new commands.~~
6. ~~Create non-interactive commands processing.~~
8. ~~Create utils for common calls~~
9. ~~Add general error handling for errors not appropriately handled in command.~~
9. Print something cool when the CLI starts in interactive mode. 
10. ~~Consider a builder pattern for creating commands. Using the builder will reduce the neccessity for code or patterns to be repeated across different commands.~~
1. Appropriately configure a linter to ensure code conforms to pattern on every commit
1. Setup git hooks when building the CLI binary to use linter
5. Figure out cleaner way to print help info for each command (using the FlagSet for each command from Main was not working)
9. Refactor Assume to set env variables and use the default credentials to assume the role (with the option of passing in a profile optional)
10. Move elasticsearch cluster list to external config
11. Clean up elastic code a bit 
12. For KSSH, make system user (i.e. ubuntu) a user-input variable with the default as Ubuntu 
13. Pull AWS Region (in utils.go) from config file or profile