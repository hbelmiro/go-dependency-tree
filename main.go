package main

import (
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func main() {
	up := flag.Bool("up", false, "show path(s) from root to the matched module")
	down := flag.Bool("down", false, "show the matched module's subtree")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [flags] [search-term]

Visualize Go module dependency trees.

Without a search term, prints the full dependency tree.
With a search term, shows paths to matching modules (--up, the default)
and/or their subtrees (--down). Both flags can be combined.

Flags:
  --up      show path(s) from root to the matched module
  --down    show the matched module's subtree

Examples:
  %s                   Print the full dependency tree
  %s spew              Show path(s) from root to modules matching "spew"
  %s --down spew       Show subtree(s) rooted at modules matching "spew"
  %s --up --down spew  Show both paths and subtrees
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}
	flag.Parse()

	lines, err := goModGraph()
	if err != nil {
		log.Fatalf("error running go mod graph [%v]", err)
	}

	nodes := readNodes(lines)
	root := getModuleName()

	target := flag.Arg(0)
	if target == "" {
		printTree(nodes, root, "", true, make(map[string]bool))
		return
	}

	if !*up && !*down {
		*up = true
	}

	paths := findPaths(nodes, root, target)
	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "no module matching %q found\n", target)
		os.Exit(1)
	}

	printed := false

	if *up {
		for _, path := range paths {
			if printed {
				fmt.Println()
			}
			printed = true
			printPath(path)
		}
	}

	if *down {
		seen := make(map[string]bool)
		for _, path := range paths {
			matched := path[len(path)-1]
			if seen[matched] {
				continue
			}
			seen[matched] = true
			if printed {
				fmt.Println()
			}
			printed = true
			printTree(nodes, matched, "", true, make(map[string]bool))
		}
	}
}

func readNodes(lines []string) map[string][]string {
	nodes := make(map[string][]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		nodeName := parts[0]
		if _, exists := nodes[nodeName]; !exists {
			nodes[nodeName] = make([]string, 0)
		}
		nodes[nodeName] = append(nodes[nodeName], parts[1])
	}
	return nodes
}

func printTree(nodes map[string][]string, root string, indent string, isLast bool, visited map[string]bool) {
	if visited[root] {
		if isLast {
			fmt.Println(indent + "└── " + root + " (cycle)")
		} else {
			fmt.Println(indent + "├── " + root + " (cycle)")
		}
		return
	}

	visited[root] = true

	if indent != "" {
		if isLast {
			fmt.Println(indent + "└── " + root)
			indent += "    "
		} else {
			fmt.Println(indent + "├── " + root)
			indent += "│   "
		}
	} else {
		fmt.Println(root)
		indent += "    "
	}

	children, exists := nodes[root]
	if !exists {
		return
	}

	for i, child := range children {
		isLastChild := i == len(children)-1
		childVisited := make(map[string]bool, len(visited))
		maps.Copy(childVisited, visited)
		printTree(nodes, child, indent, isLastChild, childVisited)
	}
}

func findPaths(nodes map[string][]string, root, target string) [][]string {
	var results [][]string
	var dfs func(node string, path []string, visited map[string]bool)
	dfs = func(node string, path []string, visited map[string]bool) {
		if visited[node] {
			return
		}
		visited[node] = true
		path = append(path, node)

		if strings.Contains(node, target) {
			result := make([]string, len(path))
			copy(result, path)
			results = append(results, result)
		}

		for _, child := range nodes[node] {
			childVisited := make(map[string]bool, len(visited))
			maps.Copy(childVisited, visited)
			dfs(child, path, childVisited)
		}
	}
	dfs(root, nil, make(map[string]bool))
	return results
}

func printPath(path []string) {
	if len(path) == 0 {
		return
	}
	fmt.Println(path[0])
	indent := "    "
	for i := 1; i < len(path); i++ {
		fmt.Println(indent + "└── " + path[i])
		indent += "    "
	}
}

func getModuleName() string {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("error running go list -m [%v]", err)
	}
	return strings.TrimSpace(string(output))
}

func goModGraph() ([]string, error) {
	output, err := exec.Command("go", "mod", "graph").Output()
	if err != nil {
		return nil, err
	}
	return slices.Collect(strings.SplitSeq(string(output), "\n")), nil
}
