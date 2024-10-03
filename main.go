package main

import (
	"flag"
	"fmt"
	"orf/object"
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

		pathArg := initCmd.Arg(0)
		_, err := repository.CreateRepo(pathArg)
		if err != nil {
			fmt.Printf("error initializing repo: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Succesfully initialized repo\n")

	case "cat":
		initCmd := flag.NewFlagSet("cat", flag.ExitOnError)
		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 2 {
			fmt.Println("expected format and object argument")
			os.Exit(1)
		}

		formatArg := initCmd.Arg(0)
		objectArg := initCmd.Arg(1)

		if !contains([]string{"blob", "commit", "tag", "tree"}, formatArg) {
			fmt.Println("incorrect format argument")
			os.Exit(1)
		}

		repo, err := repository.FindRepo(".", true)
		if err != nil {
			fmt.Printf("error finding repo: %v\n", err)
			os.Exit(1)
		}

		newObject, err := object.ReadObject(repo.WorkTree, objectArg)
		if err != nil {
			fmt.Printf("error reading object: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(newObject.Data)

	case "log":
		logCmd := flag.NewFlagSet("log", flag.ExitOnError)
		logCmd.Parse(os.Args[2:])
		fmt.Printf("Log Command\n")

	default:
		fmt.Println("Expected 'init' or 'log' subcommands")
		os.Exit(1)
	}
}

func contains(candidates []string, target string) bool {
	for _, candidate := range candidates {
		if target == candidate {
			return true
		}
	}
	return false
}
