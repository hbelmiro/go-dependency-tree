# go-dependency-tree

Visualize Go module dependency trees in the terminal.

Runs `go mod graph` in the current module and renders the result as an ASCII tree.

## Installation

```sh
go install go-dependency-tree@latest
```

## Usage

```sh
# Print the full dependency tree
go-dependency-tree

# Show path(s) from root to modules matching a search term
go-dependency-tree spew

# Show subtree(s) rooted at matching modules
go-dependency-tree --down spew

# Show both paths and subtrees
go-dependency-tree --up --down spew
```

### Flags

| Flag     | Description                                                                        |
|----------|------------------------------------------------------------------------------------|
| `--up`   | Show path(s) from root to the matched module (default when a search term is given) |
| `--down` | Show the matched module's subtree                                                  |

Both flags can be combined.

## Example output

```
my-project
    ├── github.com/foo/bar@v1.0.0
    │   └── github.com/baz/qux@v2.1.0
    └── github.com/hello/world@v0.3.0
```

## Requirements

- Go 1.26 or later
- Must be run inside a Go module (directory with `go.mod`)
