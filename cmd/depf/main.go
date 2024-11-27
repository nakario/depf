package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nakario/depf"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <file> [package patterns...]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "  <file>                The .go file to analyze.")
		fmt.Fprintln(os.Stderr, "  [package patterns...] Optional package patterns to change the search scope.")
		fmt.Fprintln(os.Stderr, "                        If not provided, the parent directory of <file> is used.")
		flag.PrintDefaults()
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		return
	}

	file := args[0]
	packagePatterns := args[1:]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get current working directory: %v\n", err)
		return
	}

	deps, err := depf.Depf(file, packagePatterns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	deps, err = getRelativePaths(cwd, deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	for _, dep := range deps {
		fmt.Println(dep)
	}
}

// getRelativePaths converts absolute paths to relative paths from cwd.
func getRelativePaths(cwd string, files []string) ([]string, error) {
	relativePaths := make([]string, 0, len(files))
	for _, f := range files {
		relPath, err := filepath.Rel(cwd, f)
		if err != nil {
			return nil, fmt.Errorf("could not get relative path: %w", err)
		}
		relativePaths = append(relativePaths, relPath)
	}
	return relativePaths, nil
}
