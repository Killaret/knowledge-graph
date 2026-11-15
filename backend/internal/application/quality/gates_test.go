package quality

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func sig(mut func(*Signals)) Signals {
	s := Signals{Kind: KindText, Words: 100}
	if mut != nil {
		mut(&s)
	}
	return s
}

func TestEvaluate_Gates(t *testing.T) {
	tests := []struct {
		name        string
		signals     Signals
		wantVerdict string
		wantGates   []string
	}{
		{"clean text", sig(nil), VerdictCreate, []string{}},
		{"stub", sig(func(s *Signals) { s.Kind = KindStub; s.Words = 0; s.HasSourceLink = true }), VerdictEnrich, []string{GateStub}},
		{"truncated", sig(func(s *Signals) { s.TruncatedByImport = true }), VerdictEnrich, []string{GateTruncated}},
		{"empty", sig(func(s *Signals) { s.Words = 0 }), VerdictManual, []string{GateEmpty}},
		{"mojibake", sig(func(s *Signals) { s.Mojibake = true }), VerdictManual, []string{GateMojibake}},
		// A stub note is empty by definition but carries a link — it is an
		// enrich case, not a manual one.
		{"stub+truncated", sig(func(s *Signals) {
			s.Kind = KindStub
			s.Words = 0
			s.HasSourceLink = true
			s.TruncatedByImport = true
		}), VerdictEnrich, []string{GateStub, GateTruncated}},
		{"mojibake beats enrich", sig(func(s *Signals) {
			s.Mojibake = true
			s.TruncatedByImport = true
		}), VerdictManual, []string{GateTruncated, GateMojibake}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verdict, reasons := Evaluate(tt.signals)
			assert.Equal(t, tt.wantVerdict, verdict)
			assert.Equal(t, tt.wantGates, reasons)
		})
	}
}

func TestSameSignals(t *testing.T) {
	a := sig(nil)
	assert.True(t, SameSignals(a, a))

	b := a
	b.Words++
	assert.False(t, SameSignals(a, b))

	b = a
	v := 0.5
	b.CoherenceMin = &v
	assert.False(t, SameSignals(a, b))
}
