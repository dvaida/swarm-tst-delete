# Plan: Issue #20 - Implement Software Project Detection

## Overview
Implement project and language detection for the swarm-indexer tool. This involves:
1. Detecting project types by marker files (go.mod, package.json, etc.)
2. Detecting VCS presence (.git, .svn, .hg)
3. Detecting IDE configs (.vscode, .idea)
4. Detecting file language by extension
5. Extracting dependencies from go.mod and package.json

## Files to Create
- `internal/detector/project.go` - Project detection logic
- `internal/detector/language.go` - Language detection by file extension
- `internal/detector/project_test.go` - Integration tests for project detection
- `internal/detector/language_test.go` - Integration tests for language detection
- `go.mod` - Go module file (project needs this)

## Integration Tests to Write

### Project Detection Tests (`project_test.go`)
1. **TestDetectProject_GoProject** - Create temp dir with go.mod, verify Type="go"
2. **TestDetectProject_NodeProject** - Create temp dir with package.json, verify Type="node"
3. **TestDetectProject_PythonProject_Requirements** - Create temp dir with requirements.txt
4. **TestDetectProject_PythonProject_Pyproject** - Create temp dir with pyproject.toml
5. **TestDetectProject_PythonProject_Setup** - Create temp dir with setup.py
6. **TestDetectProject_RustProject** - Create temp dir with Cargo.toml
7. **TestDetectProject_JavaProject_Maven** - Create temp dir with pom.xml
8. **TestDetectProject_JavaProject_Gradle** - Create temp dir with build.gradle
9. **TestDetectProject_RubyProject** - Create temp dir with Gemfile
10. **TestDetectProject_UnknownProject** - Empty temp dir, verify Type="unknown"
11. **TestDetectProject_GitVCS** - Create temp dir with .git, verify HasVCS=true, VCSType="git"
12. **TestDetectProject_SvnVCS** - Create temp dir with .svn, verify VCSType="svn"
13. **TestDetectProject_HgVCS** - Create temp dir with .hg, verify VCSType="hg"
14. **TestDetectProject_VSCodeConfig** - Create temp dir with .vscode, verify HasIDEConfig=true
15. **TestDetectProject_IdeaConfig** - Create temp dir with .idea, verify HasIDEConfig=true
16. **TestDetectProject_GoModDependencies** - Create go.mod with deps, verify extraction
17. **TestDetectProject_PackageJsonDependencies** - Create package.json with deps, verify extraction
18. **TestDetectProject_NonExistentDir** - Call with bad path, verify error returned

### Language Detection Tests (`language_test.go`)
1. **TestDetectLanguage_Go** - file.go -> "go"
2. **TestDetectLanguage_Python** - file.py -> "python"
3. **TestDetectLanguage_JavaScript** - file.js -> "javascript"
4. **TestDetectLanguage_TypeScript** - file.ts -> "typescript"
5. **TestDetectLanguage_Java** - file.java -> "java"
6. **TestDetectLanguage_Rust** - file.rs -> "rust"
7. **TestDetectLanguage_Ruby** - file.rb -> "ruby"
8. **TestDetectLanguage_C** - file.c -> "c"
9. **TestDetectLanguage_Cpp** - file.cpp -> "cpp"
10. **TestDetectLanguage_Header** - file.h -> "c"
11. **TestDetectLanguage_Markdown** - file.md -> "markdown"
12. **TestDetectLanguage_JSON** - file.json -> "json"
13. **TestDetectLanguage_YAML** - file.yaml -> "yaml", file.yml -> "yaml"
14. **TestDetectLanguage_TOML** - file.toml -> "toml"
15. **TestDetectLanguage_Unknown** - file.xyz -> "unknown"

## Implementation Approach

### Step 1: Initialize Go Module
Create go.mod for the project.

### Step 2: Write Tests First
Create test files with all integration tests that create temporary directories and files.

### Step 3: Create Stubs
Create project.go and language.go with the required interfaces, returning errors/panics.

### Step 4: Verify Tests Fail
Run `go test` and confirm all tests fail.

### Step 5: Implement
1. `language.go`: Simple extension map lookup
2. `project.go`:
   - Check for marker files to determine project type
   - Check for .git/.svn/.hg for VCS
   - Check for .vscode/.idea for IDE config
   - Parse go.mod and package.json for dependencies

### Step 6: Verify Tests Pass
Run `go test` and confirm all tests pass.
