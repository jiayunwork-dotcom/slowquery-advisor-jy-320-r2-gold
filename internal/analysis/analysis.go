// Package analysis 对慢查询条目做聚合统计并给出优化建议。
package analysis

import (
	"strings"

	"slowquery-advisor/internal/fingerprint"
	"slowquery-advisor/internal/parse"
)

// Stats 是某一指纹的聚合统计。
type Stats struct {
	Fingerprint   string
	Count         int
	TotalTime     float64
	MaxTime       float64
	AvgTime       float64
	MaxSent       int
	MaxExamined   int
	TotalExamined int
}

// Aggregate 按指纹聚合慢查询条目。
func Aggregate(entries []parse.Entry) *Result {
	a := &Result{Items: map[string]*Stats{}}
	for _, e := range entries {
		fp := fingerprint.Fingerprint(e.SQL)
		s, ok := a.Items[fp]
		if !ok {
			s = &Stats{Fingerprint: fp}
			a.Items[fp] = s
		}
		s.Count++
		s.TotalTime += e.QueryTime
		if e.RowsSent > s.MaxSent {
			s.MaxSent = e.RowsSent
		}
		s.TotalExamined += e.RowsExamined
		if e.RowsExamined > s.MaxExamined {
			s.MaxExamined = e.RowsExamined
		}
	}
	for _, s := range a.Items {
		s.AvgTime = s.TotalTime / float64(s.Count)
	}
	return a
}

// Result 是聚合结果的集合。
type Result struct {
	Items map[string]*Stats
}

// Suggest 针对聚合统计给出索引/优化建议。
func Suggest(s *Stats) []string {
	var tips []string
	if s == nil || s.Count == 0 {
		return tips
	}
	if s.MaxExamined > 1000 && s.MaxSent > 0 && s.MaxExamined/s.MaxSent > 1000 {
		tips = append(tips, "扫描行数远高于返回行数，考虑为 WHERE/JOIN 列添加索引以减少全表扫描")
	}
	lower := s.Fingerprint
	if strings.Contains(lower, "like") && strings.Contains(lower, "%") {
		tips = append(tips, "存在前置通配符 LIKE '%...%'，无法利用 B-Tree 索引，考虑全文索引或后缀通配")
	}
	if !strings.Contains(lower, "where") {
		tips = append(tips, "缺少 WHERE 条件，可能触发全表扫描，建议增加过滤条件")
	}
	return tips
}
