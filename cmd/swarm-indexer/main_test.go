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
	cmd.SetArgs([]string{"search", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "search") {
		t.Errorf("expected help output to contain 'search', got:\n%s", output)
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

func TestSearchCommand_Runs(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"search", "test query"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
