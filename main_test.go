package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestPrintTree_DiamondDependency(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"B":    {"C"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	require.NotContains(t, output, "(cycle)")

	cCount := 0
	for line := range strings.SplitSeq(output, "\n") {
		if strings.Contains(line, "C") && !strings.Contains(line, "(cycle)") {
			cCount++
		}
	}
	assert.Equal(t, 2, cCount, "C should appear under both A and B")
}

func TestPrintTree_LinearChain(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
		"A":    {"B"},
		"B":    {"C"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	assert.NotContains(t, output, "(cycle)")
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "A")
	assert.Contains(t, output, "B")
	assert.Contains(t, output, "C")
}

func TestPrintTree_RealCycle(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
		"A":    {"B"},
		"B":    {"A"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	assert.Contains(t, output, "A (cycle)")
}

func TestPrintTree_CycleOnNonLastChild(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
		"A":    {"root", "B"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    └── A", lines[1])
	assert.Equal(t, "        ├── root (cycle)", lines[2])
	assert.Equal(t, "        └── B", lines[3])
}

func TestPrintTree_SingleNode(t *testing.T) {
	nodes := map[string][]string{}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	require.Equal(t, "root\n", output)
}

func TestPrintTree_DiamondWithSubtree(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"B":    {"C"},
		"C":    {"D"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	require.NotContains(t, output, "(cycle)")

	dCount := 0
	for line := range strings.SplitSeq(output, "\n") {
		if strings.Contains(line, "D") {
			dCount++
		}
	}
	assert.Equal(t, 2, dCount, "D (child of C) should appear under both branches")
}

func TestPrintTree_DiamondWithCycle(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"B":    {"C"},
		"C":    {"A"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	cCount := 0
	for line := range strings.SplitSeq(output, "\n") {
		if strings.Contains(line, "C") && !strings.Contains(line, "(cycle)") {
			cCount++
		}
	}
	assert.Equal(t, 2, cCount, "C should appear under both A and B")
	assert.Contains(t, output, "A (cycle)", "first branch should detect A→C→A cycle")
	assert.Contains(t, output, "C (cycle)", "second branch should detect B→C→A→C cycle")
}

func TestPrintTree_BranchingCharacters(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B", "C"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    ├── A", lines[1])
	assert.Equal(t, "    ├── B", lines[2])
	assert.Equal(t, "    └── C", lines[3])
}

func TestPrintTree_IndentContinuation(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"B":    {"D"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 5)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    ├── A", lines[1])
	assert.Equal(t, "    │   └── C", lines[2])
	assert.Equal(t, "    └── B", lines[3])
	assert.Equal(t, "        └── D", lines[4])
}

func TestPrintTree_DeepIndentPropagation(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"C":    {"D"},
	}

	output := captureOutput(func() {
		printTree(nodes, "root", "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 5)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    ├── A", lines[1])
	assert.Equal(t, "    │   └── C", lines[2])
	assert.Equal(t, "    │       └── D", lines[3])
	assert.Equal(t, "    └── B", lines[4])
}

func TestFindPaths_LinearChain(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
		"A":    {"B"},
		"B":    {"C"},
	}

	paths := findPaths(nodes, "root", "C")

	assert.Equal(t, [][]string{{"root", "A", "B", "C"}}, paths)
}

func TestFindPaths_Diamond(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"B":    {"C"},
	}

	paths := findPaths(nodes, "root", "C")

	assert.Equal(t, [][]string{{"root", "A", "C"}, {"root", "B", "C"}}, paths)
}

func TestFindPaths_NoMatch(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
		"A":    {"B"},
	}

	paths := findPaths(nodes, "root", "X")

	assert.Empty(t, paths)
}

func TestFindPaths_PartialMatch(t *testing.T) {
	nodes := map[string][]string{
		"root": {"github.com/davecgh/go-spew@v1.1.1"},
	}

	paths := findPaths(nodes, "root", "spew")

	assert.Equal(t, [][]string{{"root", "github.com/davecgh/go-spew@v1.1.1"}}, paths)
}

func TestFindPaths_Cycle(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
		"A":    {"B"},
		"B":    {"A"},
	}

	paths := findPaths(nodes, "root", "B")

	assert.Equal(t, [][]string{{"root", "A", "B"}}, paths)
}

func TestFindPaths_TargetIsRoot(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A"},
	}

	paths := findPaths(nodes, "root", "root")

	assert.Equal(t, [][]string{{"root"}}, paths)
}

func TestFindPaths_ContinuesBelowMatch(t *testing.T) {
	nodes := map[string][]string{
		"root":  {"A-lib"},
		"A-lib": {"C-lib"},
	}

	paths := findPaths(nodes, "root", "lib")

	assert.Equal(t, [][]string{{"root", "A-lib"}, {"root", "A-lib", "C-lib"}}, paths)
}

func TestFindPaths_MultipleBranchesBelowMatch(t *testing.T) {
	nodes := map[string][]string{
		"root":  {"A-lib"},
		"A-lib": {"B-lib", "C-lib"},
	}

	paths := findPaths(nodes, "root", "lib")

	assert.Equal(t, [][]string{
		{"root", "A-lib"},
		{"root", "A-lib", "B-lib"},
		{"root", "A-lib", "C-lib"},
	}, paths)
}

func TestFindPaths_RootAndDescendantMatch(t *testing.T) {
	nodes := map[string][]string{
		"root-lib": {"A-lib"},
		"A-lib":    {"B"},
	}

	paths := findPaths(nodes, "root-lib", "lib")

	assert.Equal(t, [][]string{{"root-lib"}, {"root-lib", "A-lib"}}, paths)
}

func TestFindPaths_DiamondWithCycle(t *testing.T) {
	nodes := map[string][]string{
		"root": {"A", "B"},
		"A":    {"C"},
		"B":    {"C"},
		"C":    {"A"},
	}

	paths := findPaths(nodes, "root", "C")

	assert.Equal(t, [][]string{{"root", "A", "C"}, {"root", "B", "C"}}, paths)
}

func TestFindPaths_DistinctNodesMatchingSameSubstring(t *testing.T) {
	nodes := map[string][]string{
		"root": {"alpha-x", "beta-x"},
	}

	paths := findPaths(nodes, "root", "x")

	assert.Equal(t, [][]string{{"root", "alpha-x"}, {"root", "beta-x"}}, paths)
}

func TestBuildMergedTree_SinglePath(t *testing.T) {
	paths := [][]string{{"root", "A", "B", "C"}}

	merged := buildMergedTree(paths)

	assert.Equal(t, []string{"A"}, merged["root"])
	assert.Equal(t, []string{"B"}, merged["A"])
	assert.Equal(t, []string{"C"}, merged["B"])
}

func TestBuildMergedTree_DiamondPaths(t *testing.T) {
	paths := [][]string{
		{"root", "A", "C"},
		{"root", "B", "C"},
	}

	merged := buildMergedTree(paths)

	assert.Equal(t, []string{"A", "B"}, merged["root"])
	assert.Equal(t, []string{"C"}, merged["A"])
	assert.Equal(t, []string{"C"}, merged["B"])
}

func TestBuildMergedTree_SharedPrefix(t *testing.T) {
	paths := [][]string{
		{"root", "A", "B"},
		{"root", "A", "C"},
	}

	merged := buildMergedTree(paths)

	assert.Equal(t, []string{"A"}, merged["root"])
	assert.Equal(t, []string{"B", "C"}, merged["A"])
}

func TestBuildMergedTree_SingleNodePath(t *testing.T) {
	paths := [][]string{{"root"}}

	merged := buildMergedTree(paths)

	assert.Empty(t, merged)
}

func TestBuildMergedTree_ChildOrder(t *testing.T) {
	paths := [][]string{
		{"root", "X", "Z"},
		{"root", "Y", "Z"},
		{"root", "X", "W"},
	}

	merged := buildMergedTree(paths)

	assert.Equal(t, []string{"X", "Y"}, merged["root"], "children should appear in first-seen order")
	assert.Equal(t, []string{"Z", "W"}, merged["X"], "children should appear in first-seen order")
}

func TestMergedOutput_DiamondPaths(t *testing.T) {
	paths := [][]string{
		{"root", "A", "C"},
		{"root", "B", "C"},
	}

	merged := buildMergedTree(paths)
	output := captureOutput(func() {
		printTree(merged, paths[0][0], "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 5)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    ├── A", lines[1])
	assert.Equal(t, "    │   └── C", lines[2])
	assert.Equal(t, "    └── B", lines[3])
	assert.Equal(t, "        └── C", lines[4])
}

func TestMergedOutput_SinglePath(t *testing.T) {
	paths := [][]string{
		{"root", "A", "B", "C"},
	}

	merged := buildMergedTree(paths)
	output := captureOutput(func() {
		printTree(merged, paths[0][0], "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    └── A", lines[1])
	assert.Equal(t, "        └── B", lines[2])
	assert.Equal(t, "            └── C", lines[3])
}

func TestMergedOutput_SharedPrefix(t *testing.T) {
	paths := [][]string{
		{"root", "A", "B"},
		{"root", "A", "C"},
	}

	merged := buildMergedTree(paths)
	output := captureOutput(func() {
		printTree(merged, paths[0][0], "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    └── A", lines[1])
	assert.Equal(t, "        ├── B", lines[2])
	assert.Equal(t, "        └── C", lines[3])
}

func TestMergedOutput_DiamondWithSubtree(t *testing.T) {
	paths := [][]string{
		{"root", "A", "C", "D"},
		{"root", "B", "C", "D"},
	}

	merged := buildMergedTree(paths)
	output := captureOutput(func() {
		printTree(merged, paths[0][0], "", true, make(map[string]bool))
	})

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	require.Len(t, lines, 7)
	assert.Equal(t, "root", lines[0])
	assert.Equal(t, "    ├── A", lines[1])
	assert.Equal(t, "    │   └── C", lines[2])
	assert.Equal(t, "    │       └── D", lines[3])
	assert.Equal(t, "    └── B", lines[4])
	assert.Equal(t, "        └── C", lines[5])
	assert.Equal(t, "            └── D", lines[6])
}

func TestReadNodes(t *testing.T) {
	lines := []string{
		"root@v1.0 A@v1.0",
		"root@v1.0 B@v2.0",
		"A@v1.0 C@v0.1",
		"",
		"malformed-line",
	}

	nodes := readNodes(lines)

	assert.Equal(t, []string{"A@v1.0", "B@v2.0"}, nodes["root@v1.0"])
	assert.Equal(t, []string{"C@v0.1"}, nodes["A@v1.0"])
	_, hasMalformed := nodes["malformed-line"]
	assert.False(t, hasMalformed)
}
