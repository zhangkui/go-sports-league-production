# BUG-011 Reproduction

Team creation writes the team row successfully but creates the initial pending registration with team_id=0. The correct behavior is to persist exactly one pending registration linked to the newly created team and no orphan registration.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug011_BusinessRegression$'
```
