package main

import (
	"fmt"
	"io"
	"os"

	"slowquery-advisor/internal/analysis"
	"slowquery-advisor/internal/parse"
	"slowquery-advisor/internal/report"
)

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  slowquery-advisor analyze --log FILE [--format text|json] [--min-time SECONDS]")
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) < 1 || args[0] != "analyze" {
		usage(stderr)
		return fmt.Errorf("missing command")
	}
	var logPath, format string
	minTime := 0.0
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--log":
			if i+1 < len(args) {
				logPath = args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--min-time":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%f", &minTime)
				i++
			}
		}
	}
	if logPath == "" {
		usage(stderr)
		return fmt.Errorf("analyze requires --log")
	}
	f, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer f.Close()
	entries, err := parse.ParseLog(f)
	if err != nil {
		return err
	}
	agg := analysis.Aggregate(entries)
	// 按阈值过滤慢查询。
	filtered := &analysis.Result{Items: map[string]*analysis.Stats{}}
	for fp, s := range agg.Items {
		if s.MaxTime >= minTime {
			filtered.Items[fp] = s
		}
	}
	if format == "json" {
		return report.RenderJSON(filtered, stdout)
	}
	report.RenderText(filtered, stdout)
	return nil
}
