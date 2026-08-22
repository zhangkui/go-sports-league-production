# Evaluation Guide

- Repository: zhangkui/go-sports-league-production
- Branch: test_model_fix3

## Public verification

```bash
go test ./scripts/verify -count=1 -run '^TestBug003_BusinessRegression$'
```
