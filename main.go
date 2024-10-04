package main

import (
	"flag"
	"fmt"
	"orf/cmd"
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
			fmt.Println("expected format and hash argument")
			os.Exit(1)
		}

		formatArg := initCmd.Arg(0)
		hashArg := initCmd.Arg(1)

		if !contains([]string{"blob", "commit", "tag", "tree"}, formatArg) {
			fmt.Println("incorrect format argument")
			os.Exit(1)
		}

		obj, err := cmd.CatObject(hashArg)
		if err != nil {
			fmt.Printf("error returning object: %v/n", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", string(obj.Data))

	case "hash":
		initCmd := flag.NewFlagSet("hash", flag.ExitOnError)

		// Define the flags
		writeFlag := initCmd.Bool("w", false, "Write to a file")
		formatFlag := initCmd.String("format", "blob", "Specify the format (blob, commit, tag, tree)")

		// Validate the formatFlag
		if !contains([]string{"blob", "commit", "tag", "tree"}, *formatFlag) {
			fmt.Println("incorrect type argument")
			os.Exit(1)
		}

		// Parse the flags
		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 1 {
			fmt.Println("expected path argument")
			os.Exit(1)
		}

		pathArg := initCmd.Arg(0)

		hash, err := cmd.HashObject(pathArg, *formatFlag, *writeFlag)
		if err != nil {
			fmt.Printf("error writing object: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Object written with hash: %s\n", hash)

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
