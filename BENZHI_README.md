# BUG-013 Evaluation Delivery

This branch contains the BUG-013 model fix and the public evaluation harness.

## Environment

- Go module: `backend/go.mod` (`go 1.22`)
- Local verification environment: `go1.26.1 windows/amd64`
- MySQL and Redis use the public fixture defaults documented by the project.

## Build

```bash
cd backend
go build ./...
```

## Public Verification

```bash
cd backend
go test ./scripts/verify -count=1 -run '^TestBug013_BusinessRegression$'
```

## Docker

```bash
./build_benzhi_docker.sh
```

The Docker build targets `linux/amd64` and retains the Go toolchain.

## Result Evidence

- Defect baseline: the public regression command reaches the target business assertions and fails.
- Model fix commit: `8759c5ce2449ad32d503c4ee39131f017f3bca0b`.
- Post-fix verification: the same public command passes.
- Raw model trajectories are delivered separately and are not stored in this Git branch.
