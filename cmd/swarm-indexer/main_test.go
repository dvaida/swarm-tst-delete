package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestSearchCommand_Help tests that the search command shows help
func TestSearchCommand_Help(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"search", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("search --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "search") {
		t.Error("expected help output to mention 'search'")
	}
	if !strings.Contains(output, "--limit") {
		t.Error("expected help output to mention '--limit' flag")
	}
	if !strings.Contains(output, "--json") {
		t.Error("expected help output to mention '--json' flag")
	}
}

// TestSearchCommand_RequiresQuery tests that search requires a query argument
func TestSearchCommand_RequiresQuery(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"search"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no query provided")
	}
}

// TestSearchCommand_DefaultLimit tests that default limit is 10
func TestSearchCommand_DefaultLimit(t *testing.T) {
	cmd := newRootCmd()

	// Find the search command and check default limit
	searchCmd, _, err := cmd.Find([]string{"search"})
	if err != nil {
		t.Fatalf("failed to find search command: %v", err)
	}

	limitFlag := searchCmd.Flag("limit")
	if limitFlag == nil {
		t.Fatal("expected --limit flag to exist")
	}
	if limitFlag.DefValue != "10" {
		t.Errorf("expected default limit to be 10, got %s", limitFlag.DefValue)
	}
}

// TestSearchCommand_JSONFlagExists tests that --json flag exists
func TestSearchCommand_JSONFlagExists(t *testing.T) {
	cmd := newRootCmd()

	searchCmd, _, err := cmd.Find([]string{"search"})
	if err != nil {
		t.Fatalf("failed to find search command: %v", err)
	}

	jsonFlag := searchCmd.Flag("json")
	if jsonFlag == nil {
		t.Fatal("expected --json flag to exist")
	}
	if jsonFlag.DefValue != "false" {
		t.Errorf("expected default json to be false, got %s", jsonFlag.DefValue)
	}
}
