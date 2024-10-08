package main

import (
	"flag"
	"fmt"
	"orf/cmd"
	"os"
)

func main() {

	if len(os.Args) == 1 {
		cmd.Help()
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

		err := cmd.Init(pathArg)
		if err != nil {
			fmt.Printf("error initializing repo: %v/n", err)
			os.Exit(1)
		}

		fmt.Printf("Succesfully initialized repo\n")
		os.Exit(1)

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

		obj, err := cmd.CatObject(hashArg, formatArg)
		if err != nil {
			fmt.Printf("error returning object: %v/n", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", string(obj.GetData()))
		os.Exit(1)

	case "hash":
		initCmd := flag.NewFlagSet("hash", flag.ExitOnError)
		writeFlag := initCmd.Bool("w", false, "Write to a file")
		formatFlag := initCmd.String("format", "blob", "Specify the format (blob, commit, tag, tree)")

		if !contains([]string{"blob", "commit", "tag", "tree"}, *formatFlag) {
			fmt.Println("incorrect type argument")
			os.Exit(1)
		}

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
		os.Exit(1)

	case "log":
		initCmd := flag.NewFlagSet("log", flag.ExitOnError)
		initCmd.Parse(os.Args[2:])

		commitArg := "HEAD"
		if initCmd.NArg() == 1 {
			commitArg = initCmd.Arg(0)

		}

		err := cmd.Log(commitArg)
		if err != nil {
			fmt.Printf("error logging commit: %v\n", err)
			os.Exit(1)
		}
		os.Exit(1)

	case "ls-tree":
		initCmd := flag.NewFlagSet("ls-tree", flag.ExitOnError)
		recursiveFlag := initCmd.Bool("r", false, "Recurse through tree")

		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 1 {
			fmt.Println("expected tree argument")
			os.Exit(1)
		}

		treeArg := initCmd.Arg(0)

		err := cmd.ListTree(treeArg, *recursiveFlag)
		if err != nil {
			fmt.Printf("error listing tree: %v\n", err)
			os.Exit(1)
		}
		os.Exit(1)

	case "checkout":
		initCmd := flag.NewFlagSet("checkout", flag.ExitOnError)

		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 2 {
			fmt.Println("expected hash & path argument")
			os.Exit(1)
		}

		hashArg := initCmd.Arg(0)
		pathArg := initCmd.Arg(1)

		err := cmd.Checkout(hashArg, pathArg)
		if err != nil {
			fmt.Printf("error checking out hash: %v\n", err)
			os.Exit(1)
		}
		os.Exit(1)

	case "help":
		cmd.Help()
		os.Exit(1)

	default:
		fmt.Println("Expected valid subcommands")
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
