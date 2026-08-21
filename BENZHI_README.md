# BUG-002 Evaluation Delivery

This branch contains the BUG-002 model fix and the public evaluation harness.

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
go test ./scripts/verify -count=1 -run '^TestBug002_BusinessRegression$'
```

## Docker

```bash
./build_benzhi_docker.sh
```

The Docker build uses `linux/amd64`, retains the Go toolchain, downloads module dependencies, and runs `go build ./...`.

## Result Evidence

- Defect baseline: the public regression command reaches the rule initialization business assertions and fails.
- Model fix commit: `ec8816d9204289b210c2982f17bd5bfba44c4928`.
- Post-fix verification: the same public command passes.
- Raw model trajectories are delivered separately and are not stored in this Git branch.
