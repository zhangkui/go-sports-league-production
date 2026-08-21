# BUG-001 Evaluation Delivery

## Project

This branch contains the BUG-001 model fix for the Go sports league application. The backend is a Go 1.22 service backed by MySQL and Redis.

## Standard Verification

Start the project test dependencies, then run from `backend/`:

```text
go test -race ./scripts/verify -count=10 -run '^TestBug001_BusinessRegression$'
```

The fixed branch must complete all ten race-enabled runs successfully.

## Docker Build

From the repository root:

```text
bash build_benzhi_docker.sh
```

Equivalent command:

```text
docker build --platform linux/amd64 -f benzhi.Dockerfile -t go-sports-league-production-bug001:latest .
```

The image keeps the Go toolchain and builds all backend packages during image creation.

## Reproduction

See `BUG_REPRO.md` for the public business scenario, trigger conditions, incorrect behavior, expected behavior, and fixed verification command.

## Validation Evidence

`BENZHI_VALIDATION/` contains the real pre-fix red output, post-fix green output, and environment information. It does not contain prompts, model trajectories, hidden tests, or answer patches.