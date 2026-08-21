# BUG-018 Validation

Public command:

```bash
go test ./scripts/verify -count=1 -run '^TestBug018_BusinessRegression$'
```

`pre_fix.txt` contains the exact Bash tool output from the formal Pre session. `diagnosis.txt` records the audited read-only diagnosis result. The raw JSONL trajectories are uploaded separately and are not committed to this branch.

The diagnosis session made no repository changes. The workflow runner added the standard evaluation delivery files afterward.