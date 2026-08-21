# BUG-001 Validation Environment

- Validation date: 2026-08-21
- Branch: `test_model_fix1`
- Base commit: `385580e49d6a71697690a759cda9a8f197174751`
- Model fix commit: `7f24be8ba608b90997e0bb541955ff6c3dc5f850`
- Go module version: `go 1.22`
- Local Go: `go1.26.1 windows/amd64`
- MySQL test endpoint: `127.0.0.1:13307`, database `league_db`
- Redis test endpoint: `127.0.0.1:6380`
- Verification command: `go test -race ./scripts/verify -count=10 -run '^TestBug001_BusinessRegression$'`
- Pre-fix result: red from target business assertions.
- Post-fix result: green, all ten race-enabled executions passed.

The validation files contain command output only. Original model trajectories and private workflow materials are not included in the Git branch.