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
