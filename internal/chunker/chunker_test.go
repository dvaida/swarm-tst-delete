package chunker

import (
	"strings"
	"testing"
)

// Test Go files split at function boundaries
func TestChunkFile_GoFunctions(t *testing.T) {
	content := `package main

import "fmt"

func hello() {
	fmt.Println("Hello")
}

func world() {
	fmt.Println("World")
}

func main() {
	hello()
	world()
}`

	chunks, err := ChunkFile("test.go", content, "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks (3 functions), got %d", len(chunks))
	}

	// Verify we have function chunks
	foundHello := false
	foundWorld := false
	foundMain := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "func hello()") {
			foundHello = true
			if chunk.ChunkType != "function" {
				t.Errorf("expected chunk type 'function', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "func world()") {
			foundWorld = true
		}
		if strings.Contains(chunk.Content, "func main()") {
			foundMain = true
		}
	}

	if !foundHello {
		t.Error("expected to find hello function chunk")
	}
	if !foundWorld {
		t.Error("expected to find world function chunk")
	}
	if !foundMain {
		t.Error("expected to find main function chunk")
	}
}

// Test Go method on struct
func TestChunkFile_GoMethods(t *testing.T) {
	content := `package main

type Server struct {
	port int
}

func (s *Server) Start() {
	// start server
}

func (s *Server) Stop() {
	// stop server
}

func NewServer(port int) *Server {
	return &Server{port: port}
}`

	chunks, err := ChunkFile("server.go", content, "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks (3 functions/methods), got %d", len(chunks))
	}

	foundStart := false
	foundStop := false
	foundNewServer := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "func (s *Server) Start()") {
			foundStart = true
		}
		if strings.Contains(chunk.Content, "func (s *Server) Stop()") {
			foundStop = true
		}
		if strings.Contains(chunk.Content, "func NewServer") {
			foundNewServer = true
		}
	}

	if !foundStart {
		t.Error("expected to find Start method chunk")
	}
	if !foundStop {
		t.Error("expected to find Stop method chunk")
	}
	if !foundNewServer {
		t.Error("expected to find NewServer function chunk")
	}
}

// Test Python files split at def/class boundaries
func TestChunkFile_PythonDefClass(t *testing.T) {
	content := `import os

class MyClass:
    def __init__(self):
        self.value = 0

    def increment(self):
        self.value += 1

def standalone_function():
    print("Hello")

def another_function(x, y):
    return x + y`

	chunks, err := ChunkFile("test.py", content, "python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks (class + functions), got %d", len(chunks))
	}

	foundClass := false
	foundStandalone := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "class MyClass:") {
			foundClass = true
			if chunk.ChunkType != "class" {
				t.Errorf("expected chunk type 'class', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "def standalone_function()") {
			foundStandalone = true
			if chunk.ChunkType != "function" {
				t.Errorf("expected chunk type 'function', got %q", chunk.ChunkType)
			}
		}
	}

	if !foundClass {
		t.Error("expected to find MyClass chunk")
	}
	if !foundStandalone {
		t.Error("expected to find standalone_function chunk")
	}
}

// Test JavaScript files split at function declarations
func TestChunkFile_JavaScript(t *testing.T) {
	content := `const express = require('express');

function handleRequest(req, res) {
    res.send('Hello');
}

const processData = (data) => {
    return data.map(x => x * 2);
};

async function fetchData(url) {
    const response = await fetch(url);
    return response.json();
}`

	chunks, err := ChunkFile("app.js", content, "javascript")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	foundHandleRequest := false
	foundFetchData := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "function handleRequest") {
			foundHandleRequest = true
			if chunk.ChunkType != "function" {
				t.Errorf("expected chunk type 'function', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "async function fetchData") {
			foundFetchData = true
		}
	}

	if !foundHandleRequest {
		t.Error("expected to find handleRequest function chunk")
	}
	if !foundFetchData {
		t.Error("expected to find fetchData function chunk")
	}
}

// Test TypeScript files with class methods
func TestChunkFile_TypeScript(t *testing.T) {
	content := `interface User {
    name: string;
    age: number;
}

class UserService {
    private users: User[] = [];

    public addUser(user: User): void {
        this.users.push(user);
    }

    public getUser(name: string): User | undefined {
        return this.users.find(u => u.name === name);
    }
}

function createUser(name: string, age: number): User {
    return { name, age };
}`

	chunks, err := ChunkFile("user.ts", content, "typescript")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	foundClass := false
	foundCreateUser := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "class UserService") {
			foundClass = true
		}
		if strings.Contains(chunk.Content, "function createUser") {
			foundCreateUser = true
		}
	}

	if !foundClass {
		t.Error("expected to find UserService class chunk")
	}
	if !foundCreateUser {
		t.Error("expected to find createUser function chunk")
	}
}

// Test Java files with method signatures
func TestChunkFile_Java(t *testing.T) {
	content := `package com.example;

public class Calculator {
    private int result;

    public Calculator() {
        this.result = 0;
    }

    public int add(int a, int b) {
        return a + b;
    }

    public int subtract(int a, int b) {
        return a - b;
    }

    private void reset() {
        this.result = 0;
    }
}`

	chunks, err := ChunkFile("Calculator.java", content, "java")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	foundAdd := false
	foundSubtract := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "public int add(") {
			foundAdd = true
			if chunk.ChunkType != "function" {
				t.Errorf("expected chunk type 'function', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "public int subtract(") {
			foundSubtract = true
		}
	}

	if !foundAdd {
		t.Error("expected to find add method chunk")
	}
	if !foundSubtract {
		t.Error("expected to find subtract method chunk")
	}
}

// Test Markdown split at headers
func TestChunkFile_Markdown(t *testing.T) {
	content := `# Main Title

This is the introduction paragraph.

## Getting Started

Follow these steps to get started.

### Prerequisites

You need Go 1.23+

### Installation

Run go install

## Usage

Here's how to use it.

### Basic Example

Just run the command.`

	chunks, err := ChunkFile("README.md", content, "markdown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks (sections), got %d", len(chunks))
	}

	foundMainTitle := false
	foundGettingStarted := false
	foundUsage := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "# Main Title") {
			foundMainTitle = true
			if chunk.ChunkType != "header" {
				t.Errorf("expected chunk type 'header', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "## Getting Started") {
			foundGettingStarted = true
		}
		if strings.Contains(chunk.Content, "## Usage") {
			foundUsage = true
		}
	}

	if !foundMainTitle {
		t.Error("expected to find Main Title chunk")
	}
	if !foundGettingStarted {
		t.Error("expected to find Getting Started chunk")
	}
	if !foundUsage {
		t.Error("expected to find Usage chunk")
	}
}

// Test plain text split at paragraph breaks
func TestChunkFile_PlainText(t *testing.T) {
	content := `This is the first paragraph.
It has multiple lines.
All part of the same paragraph.

This is the second paragraph.
Also has multiple lines.

This is the third paragraph.
Short and simple.`

	chunks, err := ChunkFile("notes.txt", content, "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks (3 paragraphs), got %d", len(chunks))
	}

	if !strings.Contains(chunks[0].Content, "first paragraph") {
		t.Error("expected first chunk to contain first paragraph")
	}
	if !strings.Contains(chunks[1].Content, "second paragraph") {
		t.Error("expected second chunk to contain second paragraph")
	}
	if !strings.Contains(chunks[2].Content, "third paragraph") {
		t.Error("expected third chunk to contain third paragraph")
	}

	for _, chunk := range chunks {
		if chunk.ChunkType != "paragraph" {
			t.Errorf("expected chunk type 'paragraph', got %q", chunk.ChunkType)
		}
	}
}

// Test YAML split by top-level keys
func TestChunkFile_YAML(t *testing.T) {
	content := `name: myproject
version: 1.0.0

dependencies:
  - package1
  - package2

scripts:
  build: go build
  test: go test

metadata:
  author: Test User
  license: MIT`

	chunks, err := ChunkFile("config.yaml", content, "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}

	foundDependencies := false
	foundScripts := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "dependencies:") {
			foundDependencies = true
			if chunk.ChunkType != "config_key" {
				t.Errorf("expected chunk type 'config_key', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "scripts:") {
			foundScripts = true
		}
	}

	if !foundDependencies {
		t.Error("expected to find dependencies chunk")
	}
	if !foundScripts {
		t.Error("expected to find scripts chunk")
	}
}

// Test JSON split by top-level keys
func TestChunkFile_JSON(t *testing.T) {
	content := `{
  "name": "myproject",
  "version": "1.0.0",
  "dependencies": {
    "package1": "1.0",
    "package2": "2.0"
  },
  "scripts": {
    "build": "go build",
    "test": "go test"
  }
}`

	chunks, err := ChunkFile("package.json", content, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	foundDependencies := false
	foundScripts := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, `"dependencies"`) {
			foundDependencies = true
			if chunk.ChunkType != "config_key" {
				t.Errorf("expected chunk type 'config_key', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, `"scripts"`) {
			foundScripts = true
		}
	}

	if !foundDependencies {
		t.Error("expected to find dependencies chunk")
	}
	if !foundScripts {
		t.Error("expected to find scripts chunk")
	}
}

// Test TOML split by sections
func TestChunkFile_TOML(t *testing.T) {
	content := `title = "My Config"

[database]
host = "localhost"
port = 5432

[server]
host = "0.0.0.0"
port = 8080

[logging]
level = "info"
format = "json"`

	chunks, err := ChunkFile("config.toml", content, "toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}

	foundDatabase := false
	foundServer := false
	foundLogging := false
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "[database]") {
			foundDatabase = true
			if chunk.ChunkType != "config_key" {
				t.Errorf("expected chunk type 'config_key', got %q", chunk.ChunkType)
			}
		}
		if strings.Contains(chunk.Content, "[server]") {
			foundServer = true
		}
		if strings.Contains(chunk.Content, "[logging]") {
			foundLogging = true
		}
	}

	if !foundDatabase {
		t.Error("expected to find database chunk")
	}
	if !foundServer {
		t.Error("expected to find server chunk")
	}
	if !foundLogging {
		t.Error("expected to find logging chunk")
	}
}

// Test large functions are split into sub-chunks
func TestChunkFile_LargeFunction(t *testing.T) {
	// Create a function with content > 4000 characters
	var builder strings.Builder
	builder.WriteString("func largeFunction() {\n")
	for i := 0; i < 200; i++ {
		builder.WriteString("\tfmt.Println(\"This is line number ")
		builder.WriteString(strings.Repeat("X", 15))
		builder.WriteString("\")\n")
	}
	builder.WriteString("}\n")

	content := "package main\n\nimport \"fmt\"\n\n" + builder.String()

	chunks, err := ChunkFile("large.go", content, "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The large function should be split into multiple chunks
	if len(chunks) < 2 {
		t.Fatalf("expected large function to be split into multiple chunks, got %d", len(chunks))
	}

	// Verify no chunk exceeds 4000 characters
	for i, chunk := range chunks {
		if len(chunk.Content) > 4000 {
			t.Errorf("chunk %d exceeds 4000 characters: %d", i, len(chunk.Content))
		}
	}
}

// Test empty file returns empty slice
func TestChunkFile_EmptyFile(t *testing.T) {
	chunks, err := ChunkFile("empty.go", "", "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty file, got %d", len(chunks))
	}
}

// Test file without clear boundaries becomes single chunk
func TestChunkFile_NoChunkBoundaries(t *testing.T) {
	content := `Just some text
without any clear
chunk boundaries
or structure.`

	chunks, err := ChunkFile("random.txt", content, "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for content without boundaries, got %d", len(chunks))
	}

	if chunks[0].Content != content {
		t.Error("expected chunk content to match original content")
	}
}

// Test accurate line numbers
func TestChunkFile_AccurateLineNumbers(t *testing.T) {
	content := `package main

func first() {
	// line 4
}

func second() {
	// line 8
}

func third() {
	// line 12
}`

	chunks, err := ChunkFile("lines.go", content, "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}

	// Find the chunks and verify line numbers
	for _, chunk := range chunks {
		if strings.Contains(chunk.Content, "func first()") {
			if chunk.StartLine != 3 {
				t.Errorf("expected first() to start at line 3, got %d", chunk.StartLine)
			}
			if chunk.EndLine != 5 {
				t.Errorf("expected first() to end at line 5, got %d", chunk.EndLine)
			}
		}
		if strings.Contains(chunk.Content, "func second()") {
			if chunk.StartLine != 7 {
				t.Errorf("expected second() to start at line 7, got %d", chunk.StartLine)
			}
			if chunk.EndLine != 9 {
				t.Errorf("expected second() to end at line 9, got %d", chunk.EndLine)
			}
		}
		if strings.Contains(chunk.Content, "func third()") {
			if chunk.StartLine != 11 {
				t.Errorf("expected third() to start at line 11, got %d", chunk.StartLine)
			}
			if chunk.EndLine != 13 {
				t.Errorf("expected third() to end at line 13, got %d", chunk.EndLine)
			}
		}
	}
}

// Test ChunkCode directly for Go
func TestChunkCode_Go(t *testing.T) {
	content := `func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}`

	chunks, err := ChunkCode(content, "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
}

// Test ChunkText directly for markdown
func TestChunkText_Markdown(t *testing.T) {
	content := `# Title

Content here.

## Section

More content.`

	chunks, err := ChunkText(content, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
}

// Test ChunkText directly for plain text
func TestChunkText_PlainText(t *testing.T) {
	content := `Paragraph one.

Paragraph two.`

	chunks, err := ChunkText(content, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
}

// Test unknown language defaults to text chunking
func TestChunkFile_UnknownLanguage(t *testing.T) {
	content := `Some content
in an unknown format.

Another paragraph.`

	chunks, err := ChunkFile("file.xyz", content, "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still produce chunks (treating as text)
	if len(chunks) == 0 {
		t.Error("expected at least some chunks for unknown language")
	}
}

// Test whitespace-only content
func TestChunkFile_WhitespaceOnly(t *testing.T) {
	content := "   \n\n   \t\t\n   "

	chunks, err := ChunkFile("whitespace.txt", content, "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Whitespace-only should return empty or minimal chunks
	if len(chunks) > 1 {
		t.Errorf("expected 0-1 chunks for whitespace-only, got %d", len(chunks))
	}
}
