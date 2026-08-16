package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"slowquery-advisor/internal/analysis"
)

// RenderText 以可读文本输出每个指纹的统计与建议。
func RenderText(a *analysis.Result, w io.Writer) {
	fps := make([]string, 0, len(a.Items))
	for fp := range a.Items {
		fps = append(fps, fp)
	}
	sort.Strings(fps)
	for _, fp := range fps {
		s := a.Items[fp]
		fmt.Fprintf(w, "fingerprint: %s\n", fp)
		fmt.Fprintf(w, "  count=%d avg=%.3fs max=%.3fs max_sent=%d max_examined=%d\n",
			s.Count, s.AvgTime, s.MaxTime, s.MaxSent, s.MaxExamined)
		for _, tip := range analysis.Suggest(s) {
			fmt.Fprintf(w, "  - advice: %s\n", tip)
		}
	}
}

// RenderJSON 以 JSON 输出聚合结果。
func RenderJSON(a *analysis.Result, w io.Writer) error {
	type item struct {
		Fingerprint   string  `json:"fingerprint"`
		Count         int     `json:"count"`
		AvgTime       float64 `json:"avg_time"`
		MaxTime       float64 `json:"max_time"`
		MaxSent       int     `json:"max_sent"`
		MaxExamined   int     `json:"max_examined"`
		Advice        []string `json:"advice"`
	}
	out := struct {
		Items []item `json:"items"`
	}{}
	for fp, s := range a.Items {
		out.Items = append(out.Items, item{
			Fingerprint: fp, Count: s.Count, AvgTime: s.AvgTime,
			MaxTime: s.MaxTime, MaxSent: s.MaxSent, MaxExamined: s.MaxExamined,
			Advice: analysis.Suggest(s),
		})
	}
	sort.Slice(out.Items, func(i, j int) bool { return out.Items[i].Fingerprint < out.Items[j].Fingerprint })
	return json.NewEncoder(w).Encode(out)
}
