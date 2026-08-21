# BUG-017 Evaluation Delivery

This branch contains the BUG-017 model fix and the public evaluation harness.

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
go test ./scripts/verify -count=1 -run '^TestBug017_BusinessRegression$'
```

## Docker

```bash
./build_benzhi_docker.sh
```

The Docker build targets `linux/amd64` and retains the Go toolchain.

## Result Evidence

- Defect baseline: the public regression command reaches the target business assertions and fails.
- Post-fix verification: the same public command passes.
- Raw model trajectories are delivered separately and are not stored in this Git branch.
