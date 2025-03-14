package main

import (
	"fmt"
	"os"
)

func printHelp() {
	fmt.Println("Usage: dedup <source_path> <destination_path>")
	fmt.Println("\nArguments:")
	fmt.Println("  source_path       Path to the source directory or file")
	fmt.Println("  destination_path  Path to the destination directory or file")
	fmt.Println("\nDescription:")
	fmt.Println("  Compares two paths and performs deduplication operations.")
}

func validateArgs() (string, string, bool) {
	args := os.Args[1:]

	// Check if help flag is provided
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printHelp()
			return "", "", false
		}
	}

	if len(args) != 2 {
		fmt.Println("Error: Expected exactly two path arguments")
		printHelp()
		return "", "", false
	}

	return args[0], args[1], true
}

func main() {
	sourcePath, destPath, valid := validateArgs()
	if !valid {
		os.Exit(1)
	}

	fmt.Printf("Source path: %s\n", sourcePath)
	fmt.Printf("Destination path: %s\n", destPath)

	// TODO: Implement the actual deduplication logic here
	fmt.Println("Deduplication process would start here...")
}
