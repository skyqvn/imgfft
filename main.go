package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "v":
		cmdView(os.Args[2:])
	case "s":
		cmdSave(os.Args[2:])
	case "r":
		cmdRestore(os.Args[2:])
	case "h", "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "错误：未知命令 %q\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}
