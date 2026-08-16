package parse

import (
	"strings"
	"testing"
)

func TestParseEntry_OK(t *testing.T) {
	block := []string{
		"# Time: 2024-01-01T00:00:00Z",
		"# User@Host: root[root] @ localhost []",
		"# Query_time: 12.345678  Lock_time: 0.000123 Rows_sent: 10  Rows_examined: 99999",
		"SET timestamp=1704067200;",
		"SELECT * FROM orders WHERE id = 42;",
	}
	e, err := ParseEntry(block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.QueryTime != 12.345678 {
		t.Fatalf("bad query time: %v", e.QueryTime)
	}
	if e.RowsSent != 10 || e.RowsExamined != 99999 {
		t.Fatalf("bad rows: sent=%d examined=%d", e.RowsSent, e.RowsExamined)
	}
	if e.User != "root" {
		t.Fatalf("bad user: %q", e.User)
	}
	if !strings.Contains(e.SQL, "SELECT") {
		t.Fatalf("sql not captured: %q", e.SQL)
	}
}

func TestParseEntry_MissingTime(t *testing.T) {
	block := []string{
		"# User@Host: root[root] @ localhost []",
		"SELECT 1;",
	}
	if _, err := ParseEntry(block); err != ErrNoQueryTime {
		t.Fatalf("expected ErrNoQueryTime, got %v", err)
	}
}

func TestParseLog_SkipsMalformed(t *testing.T) {
	in := `# Time: t1
# Query_time: 1.0  Rows_sent: 1  Rows_examined: 1
SELECT * FROM a WHERE x=1;
# Time: t2
this block has no query time
# Time: t3
# Query_time: 2.0  Rows_sent: 2  Rows_examined: 2
SELECT * FROM b WHERE y=2;
`
	entries, err := ParseLog(strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 valid entries, got %d", len(entries))
	}
}

func TestParseLog_IncompleteNeighborDoesNotFail(t *testing.T) {
	in := `# Time: t1
# Query_time: 1.5  Rows_sent: 1  Rows_examined: 10
SELECT * FROM keep_me WHERE id=1;
# Time: t2
# User@Host: root[root] @ localhost []
SELECT * FROM broken;
# Time: t3
# Query_time: 3.0  Rows_sent: 2  Rows_examined: 20
SELECT * FROM also_keep WHERE id=2;
`
	entries, err := ParseLog(strings.NewReader(in))
	if err != nil {
		t.Fatalf("incomplete neighbor must not fail the parse: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 valid entries, got %d", len(entries))
	}
	if !strings.Contains(entries[0].SQL, "keep_me") {
		t.Fatalf("lost first valid SQL: %q", entries[0].SQL)
	}
	if !strings.Contains(entries[1].SQL, "also_keep") {
		t.Fatalf("lost second valid SQL: %q", entries[1].SQL)
	}
}

func TestParseLog_EarlierBlockNotClobbered(t *testing.T) {
	in := `# Time: t1
# Query_time: 1.0  Rows_sent: 1  Rows_examined: 1
SELECT * FROM first_table WHERE a=1;
# Time: t2
# Query_time: 2.0  Rows_sent: 2  Rows_examined: 2
SELECT * FROM second_table WHERE b=2;
`
	entries, err := ParseLog(strings.NewReader(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !strings.Contains(entries[0].SQL, "first_table") {
		t.Fatalf("first entry SQL clobbered: %q", entries[0].SQL)
	}
	if strings.Contains(entries[0].SQL, "second_table") {
		t.Fatalf("first entry picked up later SQL: %q", entries[0].SQL)
	}
	if !strings.Contains(entries[1].SQL, "second_table") {
		t.Fatalf("second entry SQL: %q", entries[1].SQL)
	}
}
