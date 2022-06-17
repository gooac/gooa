package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CLIHandler func([]string)
type CLI struct {
	commands	map[string]CLIHandler
}

func NewCLI() CLI {
	cli := CLI{map[string]CLIHandler{}}

	cli.commands["compile"] = cli.Compile
	cli.commands["watch"] = cli.Watch
	// cli.commands["init"] = cli.Init
	cli.commands["help"] = cli.Help

	return cli
}

func (self *CLI) Help(args []string) {
	print(`Gooa - The Lua Preprocessor (https://github.com/gooac/gooa)

Help:
	compile [<project>]             - Compiles the given project directory or 
	                                the current directory, if invalid

	watch [<project>]               - Watches the given directory (or current if nil) 
                                    for changes and compiles the given file when needed.

	help                            - Prints this message
`)
}

func (self *CLI) Compile(args []string) {
	if len(args) == 0 {
		args = []string{"./"}
	} else {
		args[0] = filepath.Clean(args[0]) + "/"
	}

	dir, err := os.ReadDir(args[0])

	if err != nil || dir == nil {
		fmt.Printf("Failed to read directory '%v' (%v)", args[0], err)
		return
	}

	pj, cfgerr := NewProject(args[0], args[0])

	if cfgerr {
		return
	}

	cerr := pj.Compile(args[0])

	if cerr != nil {
		fmt.Printf("Error compiling directory: %v", cerr)
	} else {
		println("Finished Compilation Successfully!")
	}
}

func (self *CLI) Watch(args []string) {
	if len(args) == 0 {
		args = []string{"./"}
	} else {
		args[0] = filepath.Clean(args[0]) + "/"
	}

	dir, err := os.ReadDir(args[0])

	if err != nil || dir == nil {
		fmt.Printf("Failed to read directory '%v' (%v)", args[0], err)
		return
	}

	abs, err := filepath.Abs(args[0])

	if err != nil {
		fmt.Printf("Failed to get absolute of '%v' (%v)", args[0], err)
	}

	pathsplit := strings.Split(strings.ReplaceAll(abs, "\\", "/"), "/")
	root := pathsplit[len(pathsplit)-1]

	pj, cfgerr := NewProject(args[0], root)

	if cfgerr {
		return
	}

	pj.Watch(args[0])
}

func (self *CLI) Init(args []string) {
	location := ""
	temp := "minimal"
	progname := "none"

	if len(args) == 0 {
		
	} else if len(args) == 1 {
		fmt.Printf("Initializing a project requires 2 arguments, the template to use and the project name.\n You can provide no arguments to create a base minimal project in the current directory.")
	} else if len(args) > 2 {
		location = args[2]
	} else {
		temp = args[0]
		progname = args[1]
	}
	
	if location != "" {
		err := os.Mkdir(location, 0757)

		if err != nil && !os.IsExist(err) {
			fmt.Printf("Failed to create directory %v (%v)", location, err)
			return
		}
	} else {
		location = progname
	}

	self.InitTemplate(filepath.Clean(location), temp, progname)
}