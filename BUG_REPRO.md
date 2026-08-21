# BUG-015 Reproduction

Replacing venue availability with distinct Monday and Wednesday windows persists only one duplicated Wednesday window. The investigation identifies where the batch values become aliased and how the duplicate-key write collapses the records; no production or test file is modified by the diagnosis model.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug015_BusinessRegression$'
```
