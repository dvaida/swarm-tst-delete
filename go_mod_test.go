package main

import (
	"os"
	"strings"
	"testing"
)

// TestGoModResolved verifies that the go.mod merge conflict has been properly resolved
func TestGoModResolved(t *testing.T) {
	// Read go.mod file
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)

	// Test that there are no conflict markers
	if strings.Contains(contentStr, "<<<<<<<") {
		t.Error("go.mod still contains conflict markers (<<<<<<<)")
	}
	if strings.Contains(contentStr, ">>>>>>>") {
		t.Error("go.mod still contains conflict markers (>>>>>>>)")
	}
	if strings.Contains(contentStr, "=======") {
		t.Error("go.mod still contains conflict markers (=======)")
	}
}

// TestGoModModuleName verifies the correct module name is used
func TestGoModModuleName(t *testing.T) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	expectedModule := "module github.com/dvaida/swarm-indexer"

	if !strings.Contains(contentStr, expectedModule) {
		t.Errorf("go.mod does not contain expected module name: %s", expectedModule)
	}

	// Ensure wrong module name is not present
	wrongModule := "module github.com/swarm-indexer/swarm-indexer"
	if strings.Contains(contentStr, wrongModule) {
		t.Errorf("go.mod contains incorrect module name: %s", wrongModule)
	}
}

// TestGoModGoVersion verifies the Go version is correctly set
func TestGoModGoVersion(t *testing.T) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	expectedGoVersion := "go 1.22.2"

	if !strings.Contains(contentStr, expectedGoVersion) {
		t.Errorf("go.mod does not contain expected Go version: %s", expectedGoVersion)
	}
}

// TestGoModCobraDependency verifies cobra dependency is present
func TestGoModCobraDependency(t *testing.T) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	expectedDep := "github.com/spf13/cobra"

	if !strings.Contains(contentStr, expectedDep) {
		t.Errorf("go.mod does not contain expected cobra dependency: %s", expectedDep)
	}
}

// TestGoModSyntaxValid verifies the go.mod file has valid syntax
func TestGoModSyntaxValid(t *testing.T) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Basic syntax check - should have module and go lines
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	hasModule := false
	hasGo := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			hasModule = true
		}
		if strings.HasPrefix(line, "go ") {
			hasGo = true
		}
	}

	if !hasModule {
		t.Error("go.mod is missing module declaration")
	}
	if !hasGo {
		t.Error("go.mod is missing go version declaration")
	}
}