package depf

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/tools/go/packages"
)

// Depf finds all dependent .go files of file in the same package.
// If packagePatterns is provided, it searches dependencies in the provided packages.
// It returns a sorted list of dependent files.
// The given file is always included in the result even if there is no identifier in file.
func Depf(file string, packagePatterns ...string) ([]string, error) {
	if err := validateFile(file); err != nil {
		return nil, err
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path of file: %w", err)
	}

	pkgs, err := loadPackages(absFile, packagePatterns...)
	if err != nil {
		return nil, fmt.Errorf("could not load packages: %w", err)
	}

	dependentFiles := findDependentFilesRecursively(absFile, pkgs)

	return sortMapKeys(dependentFiles), nil
}

// validateFile checks if file is a valid .go file
func validateFile(file string) error {
	// file must be a .go file
	if len(file) < 3 || file[len(file)-3:] != ".go" {
		return fmt.Errorf("file must be a .go file")
	}
	// file must not be a directory
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("could not access file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("file must not be a directory")
	}
	return nil
}

// loadPackages loads packages in packagePatterns.
// If packagePatterns is empty, it loads the package in the directory of file.
func loadPackages(file string, packagePatterns ...string) ([]*packages.Package, error) {
	dir := filepath.Dir(file)

	if len(packagePatterns) == 0 {
		packagePatterns = append(packagePatterns, dir)
	}

	loadMode := packages.NeedName |
		packages.NeedSyntax |
		packages.NeedDeps |
		packages.NeedTypes |
		packages.NeedTypesInfo

	cfg := &packages.Config{
		Mode: loadMode,
		Dir:  dir,
	}

	return packages.Load(cfg, packagePatterns...)
}

// getFile finds the *ast.File of file in pkgs. If not found, it returns nil
func getFile(file string, pkgs []*packages.Package) *ast.File {
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			if pkg.Fset.Position(f.Pos()).Filename == file {
				return f
			}
		}
	}
	return nil
}

// findDependentFiles finds all dependent .go files of astFile.
// It iterates over all identifiers in astFile and checks if they are
// defined in pkgs, if so, it adds the file containing the definition to the result.
func findDependentFiles(astFile *ast.File, pkgs []*packages.Package) []string {
	dependentFiles := make(map[string]struct{})

	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch n := n.(type) {
		case *ast.Ident:
			for _, pkg := range pkgs {
				if obj := pkg.TypesInfo.ObjectOf(n); obj != nil {
					pos := pkg.Fset.Position(obj.Pos())
					dependentFiles[pos.Filename] = struct{}{}
				}
			}
		}

		return true
	})

	files := make([]string, 0, len(dependentFiles))
	for f := range dependentFiles {
		if getFile(f, pkgs) == nil {
			continue
		}
		files = append(files, f)
	}
	return files
}

// findDependentFilesRecursively finds all dependent .go files of file recursively.
func findDependentFilesRecursively(file string, pkgs []*packages.Package) map[string]struct{} {
	dependentFiles := make(map[string]struct{})

	queue := []string{file}
	for len(queue) > 0 {
		file := queue[0]
		queue = queue[1:]

		// queue may contain duplicate files
		if _, ok := dependentFiles[file]; ok {
			continue
		}
		dependentFiles[file] = struct{}{}

		astFile := getFile(file, pkgs)
		if astFile == nil {
			continue
		}

		dependentFilesInFile := findDependentFiles(astFile, pkgs)

		for _, f := range dependentFilesInFile {
			if _, ok := dependentFiles[f]; !ok {
				queue = append(queue, f)
			}
		}
	}

	return dependentFiles
}

func sortMapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
