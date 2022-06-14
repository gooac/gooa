package main

import (
	"os"
)

func main() {
    cli := NewCLI()
    args := os.Args[1:]

    if len(args) == 0 {
        cli.commands["help"](args)
        return
    }

    val, valid := cli.commands[args[0]]

    if valid {
        val(args[1:])
        return
    }

    cli.commands["help"](args[1:])
}