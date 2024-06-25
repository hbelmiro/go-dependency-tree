package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	lines, err := goModGraph()

	if err != nil {
		log.Fatalf("error running go mod graph [%v]", err)
	}

	nodes := readNodes(lines)
	printTree(nodes, getModuleName(), "", true, make(map[string]bool))
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
		fmt.Println(indent + "└── " + root + " (cycle)")
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
		printTree(nodes, child, indent, isLastChild, visited)
	}
}

func getModuleName() string {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running go mod graph:", err)
		return ""
	}
	return strings.TrimSpace(string(output))
}

func goModGraph() ([]string, error) {
	output, err := exec.Command("go", "mod", "graph").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	return lines, nil
}
