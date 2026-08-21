# BUG-015 Evaluation Delivery

This branch contains the BUG-015 diagnosis delivery and the public evaluation harness.

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
go test ./scripts/verify -count=1 -run '^TestBug015_BusinessRegression$'
```

## Docker

```bash
./build_benzhi_docker.sh
```

The Docker build targets `linux/amd64` and retains the Go toolchain.

## Result Evidence

- Defect baseline: the public regression command reaches the target business assertions and fails.
- Diagnosis Session made zero project changes.
- The public command remains red by design because Diagnosis does not apply a fix.
- Raw model trajectories are delivered separately and are not stored in this Git branch.
