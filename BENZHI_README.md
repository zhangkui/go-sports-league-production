# BUG-010 Evaluation Delivery

This branch contains the BUG-010 model fix and the public evaluation harness.

## Environment

- Go module: `backend/go.mod` (`go 1.22`)
- Local verification environment: `go1.26.1 windows/amd64`
- MySQL uses the public fixture default documented by the verification test.

## Build

```bash
cd backend
go build ./...
```

## Public Verification

```bash
cd backend
go test ./scripts/verify -count=1 -run '^TestBug010_BusinessRegression$'
```

## Docker

```bash
./build_benzhi_docker.sh
```

The Docker build uses `linux/amd64`, retains the Go toolchain, downloads module dependencies, and runs `go build ./...`.

## Result Evidence

- Defect baseline: the public regression command reaches the active-rule business assertions and fails.
- Model fix commit: `194c1fe3b6f5ad13bf4715ddbd35352e270afbd0`.
- Post-fix verification: the same public command passes.
- Raw model trajectories are delivered separately and are not stored in this Git branch.
