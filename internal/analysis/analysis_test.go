package analysis

import (
	"testing"

	"slowquery-advisor/internal/parse"
)

func entries() []parse.Entry {
	return []parse.Entry{
		{QueryTime: 5, RowsSent: 1, RowsExamined: 5000, SQL: "SELECT * FROM a WHERE id=1"},
		{QueryTime: 6, RowsSent: 1, RowsExamined: 6000, SQL: "SELECT * FROM a WHERE id=2"},
		{QueryTime: 1, RowsSent: 50, RowsExamined: 50, SQL: "SELECT * FROM b"},
	}
}

func TestAggregate_GroupsByFingerprint(t *testing.T) {
	agg := Aggregate(entries())
	if len(agg.Items) != 2 {
		t.Fatalf("expected 2 fingerprints, got %d", len(agg.Items))
	}
	for fp, s := range agg.Items {
		if s.Count == 2 {
			if s.MaxExamined != 6000 {
				t.Fatalf("bad max examined: %d", s.MaxExamined)
			}
			if s.AvgTime < 5.4 || s.AvgTime > 5.6 {
				t.Fatalf("bad avg time: %v", s.AvgTime)
			}
			_ = fp
		}
	}
}

func TestSuggest_FullScan(t *testing.T) {
	s := &Stats{Count: 3, MaxSent: 1, MaxExamined: 99999, Fingerprint: "select * from t where x = ?"}
	tips := Suggest(s)
	if len(tips) == 0 {
		t.Fatalf("expected at least one advice for full scan")
	}
}

func TestAggregate_Empty(t *testing.T) {
	agg := Aggregate(nil)
	if agg.Items == nil {
		t.Fatalf("Items map must be initialized, not nil")
	}
	if len(agg.Items) != 0 {
		t.Fatalf("expected empty, got %d", len(agg.Items))
	}
	if got := Suggest(&Stats{}); got != nil {
		t.Fatalf("expected nil advice for zero-count stats, got %v", got)
	}
}

func TestSuggest_NilStats(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Suggest(nil) panicked: %v", rec)
		}
	}()
	got := Suggest(nil)
	if len(got) != 0 {
		t.Fatalf("expected no advice for nil stats, got %v", got)
	}
}

func TestAggregate_TracksMaxTime(t *testing.T) {
	agg := Aggregate([]parse.Entry{
		{QueryTime: 1.2, SQL: "SELECT 1"},
		{QueryTime: 4.5, SQL: "SELECT 1"},
		{QueryTime: 3.0, SQL: "SELECT 1"},
	})
	if len(agg.Items) != 1 {
		t.Fatalf("expected 1 fingerprint, got %d", len(agg.Items))
	}
	for _, s := range agg.Items {
		if s.MaxTime != 4.5 {
			t.Fatalf("max time: got %v want 4.5", s.MaxTime)
		}
	}
}
