# BUG-008 赛季状态跃迁破坏生命周期和规则版本

创建一个处于 `draft` 状态且当前规则版本为 1 的赛季，然后请求直接变更为 `completed`。缺陷态会错误接受非法跨级跃迁，并把持久化的 `current_rule_version` 重置为 0；执行合法的 `draft` 到 `registration` 跃迁时也会错误清零规则版本。

正确结果是非法跨级跃迁被拒绝且数据库状态完全不变；合法生命周期跃迁可以成功，但不得修改当前规则版本等无关字段。

复现命令：

```text
go test ./scripts/verify -count=1 -run '^TestBug008_BusinessRegression$'
```
