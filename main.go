package main

import (
	"flag"
	"fmt"
	"orf/repository"
	"os"
)

func main() {

	if len(os.Args) == 1 {
		fmt.Println("Invoke <help> subcommand")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Expected 'init' or 'log' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 1 {
			fmt.Println("expected path argument")
			os.Exit(1)
		}

		path := initCmd.Arg(0)
		_, err := repository.CreateRepo(path)
		if err != nil {
			fmt.Printf("error initializing repo: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Succesfully initialized repo\n")

	case "log":
		logCmd := flag.NewFlagSet("log", flag.ExitOnError)
		logCmd.Parse(os.Args[2:])
		fmt.Printf("Log Command\n")

	default:
		fmt.Println("Expected 'init' or 'log' subcommands")
		os.Exit(1)
	}
}
