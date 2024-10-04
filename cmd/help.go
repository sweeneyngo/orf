package cmd

import "github.com/fatih/color"

func Help() {
	yellow := color.New(color.FgYellow).PrintfFunc()
	green := color.New(color.FgGreen).PrintfFunc()
	bold := color.New(color.Bold).PrintfFunc()
	boldYellow := color.New(color.FgYellow, color.Bold).PrintfFunc()

	green("Usage: orf [command]\n")
	bold("Available commands:\n")
	yellow("•  init <path>            Initialize a new repository at the specified path\n")
	yellow("•  cat <format> <hash>    Display the content of an object with the given hash in the specified format (blob, commit, tag, tree)\n")
	yellow("•  hash [flag] <path>     Compute the hash of the object at the specified path\n")
	boldYellow("   Options for hash:\n")
	yellow("    • -w                  Write the object to the object directory\n")
	yellow("    • --format <format>   Specify the format (i.e. blob, commit, tag, tree)\n")
	yellow("•  help                   Print all available commands\n")
}
