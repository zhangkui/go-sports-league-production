# BUG-014 Reproduction

Two reviewers concurrently approving the same transfer can both report success and can leave an incorrect final team assignment. Exactly one approval may succeed, the final transfer state must be approved once, and the player must belong to the destination team.

Reproduction command:

```text
go test -race ./scripts/verify -count=10 -run '^TestBug014_BusinessRegression$'
```
