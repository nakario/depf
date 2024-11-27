# depf

`depf` is a tool to find dependent Go files of the given file to give them to GitHub Copilot Edits as a working set.

## Installation

```sh
go install github.com/nakario/depf/cmd/depf@latest
```

## Usage

```sh
$ depf --help
Usage: depf [flags] <file> [package patterns...]
  <file>                The .go file to analyze.
  [package patterns...] Optional package patterns to change the search scope.
                        If not provided, the parent directory of <file> is used.
```

### Example

```sh
$ depf ./foo.go ./...
bar.go
baz/baz.go
foo.go
$ code -n $(depf ./foo.go ./...)
```

This opens foo.go, bar.go and baz/baz.go in a new VS Code window.
You can run GitHub Copilot Edits using all of these files in the following steps.

1. Open Copilot Edits
1. Click "Add Files..."
1. Select "Open Editors"

## License

This project is licensed under the MIT License.
