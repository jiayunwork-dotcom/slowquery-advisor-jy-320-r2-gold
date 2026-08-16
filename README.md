# slowquery-advisor

解析 MySQL 慢查询日志，按 SQL 指纹聚合，并基于扫描行比、通配符、缺失 WHERE 等启发式给出索引与优化建议。

## 用法

```bash
# 文本报告
slowquery-advisor analyze --log example/slow.log

# JSON 报告
slowquery-advisor analyze --log example/slow.log --format json

# 仅关注超过 3 秒的查询
slowquery-advisor analyze --log example/slow.log --min-time 3
```

不完整的日志块会被跳过，保证解析鲁棒；日志文件缺失时返回受控错误（退出码非 0）。

## 构建

```bash
go build ./...
go test ./...
```
