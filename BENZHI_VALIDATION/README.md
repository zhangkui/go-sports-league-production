# BUG-002 Validation

Public command:

```bash
go test ./scripts/verify -count=1 -run '^TestBug002_BusinessRegression$'
```

`pre_fix.txt` is the exact Bash tool output from the formal Pre session.
`post_fix.txt` is the exact Bash tool output from the formal Post-Pre session.

The existing public assertions were not modified by the model fix.
