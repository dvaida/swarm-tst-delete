package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()

	// Should show index subcommand
	if !strings.Contains(output, "index") {
		t.Errorf("expected help output to contain 'index', got:\n%s", output)
	}

	// Should show search subcommand
	if !strings.Contains(output, "search") {
		t.Errorf("expected help output to contain 'search', got:\n%s", output)
	}

	// Should show status subcommand
	if !strings.Contains(output, "status") {
		t.Errorf("expected help output to contain 'status', got:\n%s", output)
	}
}

func TestIndexCommand_Help(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"index", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "index") {
		t.Errorf("expected help output to contain 'index', got:\n%s", output)
	}
}

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

func TestStatusCommand_Help(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"status", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "status") {
		t.Errorf("expected help output to contain 'status', got:\n%s", output)
	}
}

func TestIndexCommand_Runs(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"index", "/tmp"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

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

func TestStatusCommand_Runs(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
