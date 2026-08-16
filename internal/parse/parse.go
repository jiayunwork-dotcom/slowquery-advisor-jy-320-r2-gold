// Package parse 解析 MySQL 慢查询日志（slow query log）。
package parse

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Entry 是单条慢查询记录。
type Entry struct {
	Time         string
	User         string
	QueryTime    float64
	LockTime     float64
	RowsSent     int
	RowsExamined int
	SQL          string
}

// ErrNoQueryTime 表示块中缺少 # Query_time: 行，视为不完整记录。
var ErrNoQueryTime = errors.New("entry missing Query_time")

// ParseLog 读取整个慢查询日志，跳过不完整的块，返回可分析的条目。
func ParseLog(r io.Reader) ([]Entry, error) {
	var out []Entry
	for _, blk := range splitBlocks(r) {
		e, err := ParseEntry(blk)
		if err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

// ParseEntry 解析单个日志块（以 # Time: 为边界的一组行）。
func ParseEntry(lines []string) (Entry, error) {
	var e Entry
	found := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "# Time:"):
			e.Time = strings.TrimSpace(line[len("# Time:"):])
		case strings.HasPrefix(line, "# User@Host:"):
			user := strings.TrimSpace(line[len("# User@Host:"):])
			if i := strings.Index(user, "["); i >= 0 {
				user = user[:i]
			}
			e.User = strings.TrimSpace(user)
		case strings.HasPrefix(line, "# Query_time:"):
			found = true
			fields := strings.Fields(line)
			m := map[string]string{}
			for i := 1; i+1 < len(fields); i += 2 {
				k := strings.TrimSuffix(fields[i], ":")
				m[k] = fields[i+1]
			}
			if v, ok := m["Query_time"]; ok {
				e.QueryTime, _ = strconv.ParseFloat(v, 64)
			}
			if v, ok := m["Lock_time"]; ok {
				e.LockTime, _ = strconv.ParseFloat(v, 64)
			}
			if v, ok := m["Rows_sent"]; ok {
				e.RowsSent, _ = strconv.Atoi(v)
			}
			if v, ok := m["Rows_examined"]; ok {
				e.RowsExamined, _ = strconv.Atoi(v)
			}
		default:
			if e.SQL == "" {
				e.SQL = line
			} else {
				e.SQL += " " + line
			}
		}
	}
	if !found {
		return Entry{}, ErrNoQueryTime
	}
	return e, nil
}

// splitBlocks 按 "# Time:" 边界把日志切分为块；首块无 Time 行也保留。
func splitBlocks(r io.Reader) [][]string {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	var blocks [][]string
	var cur []string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "# Time:") && len(cur) > 0 {
			blocks = append(blocks, cur)
			cur = nil
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		blocks = append(blocks, cur)
	}
	return blocks
}
