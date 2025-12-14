# Issue #0: Resolve merge conflicts

## Overview
Resolve merge conflicts in `go.mod` that occurred during a merge operation. The conflict is between:
- HEAD: `module github.com/dvaida/swarm-indexer` with Go 1.22.2 and cobra dependency
- origin/a/misc/issue-21-fc1b27: `module github.com/swarm-indexer/swarm-indexer` with Go 1.22

## Merge Conflict Analysis
The conflict involves:
1. **Module name**: Two different module names
   - HEAD: `github.com/dvaida/swarm-indexer`
   - Incoming: `github.com/swarm-indexer/swarm-indexer`
2. **Go version**: Different patch versions
   - HEAD: `go 1.22.2`
   - Incoming: `go 1.22`
3. **Dependencies**: HEAD has cobra dependencies, incoming branch is minimal

## Resolution Strategy
Based on the project context in CLAUDE.md, the correct module name should be `github.com/dvaida/swarm-indexer` since this appears to be the main development repository.

The resolution will:
1. Use the module name from HEAD: `github.com/dvaida/swarm-indexer`
2. Use the more specific Go version: `go 1.22.2`
3. Keep the cobra dependencies from HEAD since they're required for the CLI tool
4. Remove merge conflict markers

## Integration Tests
Since this is a merge conflict resolution (not new functionality), no new integration tests are needed. However, we should verify:
1. The go.mod file is valid Go module syntax
2. Dependencies can be resolved properly
3. The module can be built successfully

## Implementation Steps
1. Edit go.mod to resolve conflicts by choosing appropriate values
2. Verify the module syntax is correct
3. Test that go mod tidy works without errors
4. Commit the resolution