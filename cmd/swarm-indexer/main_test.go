package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()

	// Check that all three subcommands are shown
	if !strings.Contains(output, "index") {
		t.Error("Help output should contain 'index' subcommand")
	}
	if !strings.Contains(output, "search") {
		t.Error("Help output should contain 'search' subcommand")
	}
	if !strings.Contains(output, "status") {
		t.Error("Help output should contain 'status' subcommand")
	}
}

func TestIndexCommand_Exists(t *testing.T) {
	cmd := NewRootCmd()
	indexCmd, _, err := cmd.Find([]string{"index"})
	if err != nil {
		t.Fatalf("Find('index') returned error: %v", err)
	}
	if indexCmd.Name() != "index" {
		t.Errorf("Command name = %q, want %q", indexCmd.Name(), "index")
	}
}

func TestSearchCommand_Exists(t *testing.T) {
	cmd := NewRootCmd()
	searchCmd, _, err := cmd.Find([]string{"search"})
	if err != nil {
		t.Fatalf("Find('search') returned error: %v", err)
	}
	if searchCmd.Name() != "search" {
		t.Errorf("Command name = %q, want %q", searchCmd.Name(), "search")
	}
}

func TestStatusCommand_Exists(t *testing.T) {
	cmd := NewRootCmd()
	statusCmd, _, err := cmd.Find([]string{"status"})
	if err != nil {
		t.Fatalf("Find('status') returned error: %v", err)
	}
	if statusCmd.Name() != "status" {
		t.Errorf("Command name = %q, want %q", statusCmd.Name(), "status")
	}
}
