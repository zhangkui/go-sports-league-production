# BUG-018 Evaluation Delivery

This branch contains the BUG-018 public evaluation harness and diagnosis delivery files. The diagnosis model session made no Go production or test changes.

## Environment

- Go module: `backend/go.mod` (`go 1.22`)
- Local verification environment: `go1.26.1 windows/amd64`
- MySQL uses the public fixture default at `127.0.0.1:13307`.

## Build

```bash
cd backend
go build ./...
```

## Public Verification

```bash
cd backend
go test ./scripts/verify -count=1 -run '^TestBug018_BusinessRegression$'
```

## Docker

```bash
./build_benzhi_docker.sh
```

The Docker build targets `linux/amd64` and retains the Go toolchain.

## Result Evidence

- Defect baseline: the public regression command reaches the nullable-description conflict visibility assertions and fails.
- Diagnosis session: read-only production call-chain inspection; no repository files changed.
- Raw model trajectories are delivered separately and are not stored in this Git branch.