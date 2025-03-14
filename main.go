package main

import (
	"fmt"
	"os"
	"sync"
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

type fileMetadata struct {
	size int64
}

func (fm fileMetadata) equals(other fileMetadata) bool {
	return fm.size == other.size
}

func getFiles(path string) (map[string]fileMetadata, error) {
	fileMap := make(map[string]fileMetadata)

	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error accessing path %s: %w", path, err)
	}

	// If it's a single file, just return its metadata
	if !fileInfo.IsDir() {
		if fileInfo.Mode().IsRegular() {
			fileMap[path] = fileMetadata{size: fileInfo.Size()}
		}
		return fileMap, nil
	}

	// If it's a directory, process its contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", path, err)
	}

	// Process each entry in the directory
	for _, entry := range entries {
		// Skip directories and non-regular files
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			fmt.Printf("Warning: Could not get info for %s: %v\n", entry.Name(), err)
			continue
		}

		// Only include regular files
		if info.Mode().IsRegular() {
			fileMap[entry.Name()] = fileMetadata{size: info.Size()}
		}
	}

	return fileMap, nil
}

type duplicate struct {
	source      string
	destination string
}

func findDuplicates(sourceFiles, destFiles map[string]fileMetadata) []duplicate {
	var duplicates []duplicate

	for sourceName, sourceMetadata := range sourceFiles {
		if destMetadata, exists := destFiles[sourceName]; exists {
			if sourceMetadata.equals(destMetadata) {
				duplicates = append(duplicates, duplicate{
					source:      sourceName,
					destination: sourceName,
				})
			}
		}
	}

	return duplicates
}

func replaceWithSymlink(dup duplicate, sourcePath, destPath string) error {
	sourceFilePath := fmt.Sprintf("%s/%s", sourcePath, dup.source)
	destFilePath := fmt.Sprintf("%s/%s", destPath, dup.destination)
	// Validate that both files exist before proceeding
	_, err := os.Stat(sourceFilePath)
	if err != nil {
		return fmt.Errorf("source file %s does not exist: %w", sourceFilePath, err)
	}

	_, err = os.Stat(destFilePath)
	if err != nil {
		return fmt.Errorf("destination file %s does not exist: %w", destFilePath, err)
	}

	err = os.Remove(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to remove destination file %s: %w", destFilePath, err)
	}

	err = os.Symlink(sourceFilePath, destFilePath)
	if err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", destFilePath, sourceFilePath, err)
	}

	return nil
}

func main() {
	sourcePath, destPath, valid := validateArgs()
	if !valid {
		os.Exit(1)
	}

	fmt.Printf("Source path: %s\n", sourcePath)
	fmt.Printf("Destination path: %s\n", destPath)

	// Get metadata for source and destination files concurrently
	var sourceFiles, destFiles map[string]fileMetadata
	var sourceErr, destErr error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		sourceFiles, sourceErr = getFiles(sourcePath)
	}()

	go func() {
		defer wg.Done()
		destFiles, destErr = getFiles(destPath)
	}()

	wg.Wait()

	// Check for errors
	if sourceErr != nil {
		fmt.Printf("Error: %v\n", sourceErr)
		os.Exit(1)
	}

	if destErr != nil {
		fmt.Printf("Error: %v\n", destErr)
		os.Exit(1)
	}

	// Display file counts
	fmt.Printf("Found %d files in source path\n", len(sourceFiles))
	fmt.Printf("Found %d files in destination path\n", len(destFiles))

	var duplicates = findDuplicates(sourceFiles, destFiles)
	fmt.Printf("Found %d duplicates\n", len(duplicates))

	// Example of how to use the new function (commented out for safety)
	for _, dup := range duplicates {
		err := replaceWithSymlink(dup, sourcePath, destPath)
		if err != nil {
			fmt.Printf("Error replacing with symlink: %v\n", err)
		} else {
			fmt.Printf("Replaced %s with symlink to %s\n", dup.destination, dup.source)
		}
	}
}
