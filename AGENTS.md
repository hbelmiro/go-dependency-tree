# Agent Guide

Instructions for AI coding agents working on this project.

## Commands

```sh
go build ./...           # Build
go test ./...            # Test
golangci-lint run        # Lint (CI uses defaults, no config file)
```

## Conventions

- Error handling for external commands: `log.Fatalf("error running <command> [%v]", err)`.
- Test naming: `TestFunctionName_Scenario` (e.g. `TestPrintTree_DiamondDependency`).
- Tests capture stdout with the `captureOutput()` helper and assert with testify `assert`/`require`.
- Go 1.26 features are in use: `maps.Copy`, `strings.SplitSeq`, `slices.Collect`.
- Keep this file up to date when conventions or commands change.
