# BUG-013 Reproduction

Creating a player with a shirt number already used by the same team is rejected by the database but exposed with the wrong error classification. The correct behavior is a conflict response, no returned player, and exactly one persisted player with that team and number.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug013_BusinessRegression$'
```
