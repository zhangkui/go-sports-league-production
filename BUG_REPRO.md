# BUG-016 Reproduction

A partial venue update containing only status=maintenance overwrites omitted name, address, capacity, and sport fields with zero values. Omitted fields must remain unchanged while the explicitly provided status is updated.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug016_BusinessRegression$'
```
