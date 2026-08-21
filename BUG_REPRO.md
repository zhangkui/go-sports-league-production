# BUG-019 Reproduction

Canceling a schedule update request still returns a successful schedule object and persists the requested venue, date, time, and `postponed` status. A canceled request must instead return `context.Canceled`, leave the stored schedule unchanged, and preserve the number of schedule rows.

Reproduction command:

```text
go test ./scripts/verify -count=1 -run '^TestBug019_BusinessRegression$'
```
