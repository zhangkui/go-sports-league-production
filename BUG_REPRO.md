# BUG-017 Reproduction

Generating a single round-robin schedule for four approved teams returns six matches, but the pairings contain duplicates and the persisted schedules collapse into the final round. Six-team single round-robin and four-team double round-robin generation show the same instability. Each round must retain independent pairings, every team pair must occur exactly once per leg, and persisted schedules must preserve the generated round numbers.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug017_BusinessRegression$'
```
