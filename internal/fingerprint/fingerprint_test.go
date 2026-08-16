package fingerprint

import (
	"strings"
	"testing"
)

func TestFingerprint_NormalizesLiterals(t *testing.T) {
	a := Fingerprint("SELECT * FROM t WHERE id = 42 AND name = 'bob'")
	b := Fingerprint("select  *  from  t  where  id=7  and  name='alice'")
	if a != b {
		t.Fatalf("expected identical fingerprint, got %q vs %q", a, b)
	}
	if a != "select * from t where id=? and name=?" {
		t.Fatalf("unexpected fingerprint: %q", a)
	}
}

func TestFingerprint_IgnoresCaseAndSpace(t *testing.T) {
	a := Fingerprint("SELECT   a,b   FROM t")
	b := Fingerprint("select a, b from t")
	if a != b {
		t.Fatalf("expected identical fingerprint, got %q vs %q", a, b)
	}
}

func TestFingerprint_CollapsesNumericLiterals(t *testing.T) {
	a := Fingerprint("SELECT * FROM t WHERE id = 42")
	b := Fingerprint("SELECT * FROM t WHERE id = 7")
	if a != b {
		t.Fatalf("numeric literals should collapse to same fingerprint, got %q vs %q", a, b)
	}
	if !strings.Contains(a, "id=?") {
		t.Fatalf("expected id=? placeholder, got %q", a)
	}
}
