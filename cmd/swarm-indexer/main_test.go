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
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()

	// Should show the main description
	if !strings.Contains(output, "swarm-indexer") {
		t.Error("Help output should contain 'swarm-indexer'")
	}

	// Should list all subcommands
	if !strings.Contains(output, "index") {
		t.Error("Help output should list 'index' subcommand")
	}
	if !strings.Contains(output, "search") {
		t.Error("Help output should list 'search' subcommand")
	}
	if !strings.Contains(output, "status") {
		t.Error("Help output should list 'status' subcommand")
	}
}

func TestIndexCommand_Exists(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"index", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "index") {
		t.Error("Index command help should contain 'index'")
	}
}

func TestSearchCommand_Exists(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"search", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "search") {
		t.Error("Search command help should contain 'search'")
	}
}

func TestStatusCommand_Exists(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"status", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "status") {
		t.Error("Status command help should contain 'status'")
	}
}
