# BUG-018 Reproduction

The schedule conflict list omits persisted conflicts whose optional `description` is `NULL`. A conflict with a description remains visible, while team, round, or window conflicts without a description disappear from the API result even though their database rows remain present. Conflict visibility and type must not depend on whether the optional description is populated, and season filtering must continue to isolate records correctly.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug018_BusinessRegression$'
```