package quality

import "time"

// Gate names — recorded in reasons[] and surfaced to the UI.
const (
	GateStub      = "stub"
	GateTruncated = "truncated"
	GateEmpty     = "empty"
	GateMojibake  = "mojibake"
)

// MaxAutoAttempts bounds automatic assessments per text version
// (source_hash). A manual trigger does not count.
const MaxAutoAttempts = 3

// Trigger values recorded in quality_log.
const (
	TriggerAuto   = "auto"
	TriggerManual = "manual"
)

// PipelineVersion of the stage-1 quality pipeline.
const PipelineVersion = "quality-v1"

// Record is one quality assessment: signals, matched gates, the verdict and
// bookkeeping for the stop rule.
type Record struct {
	Signals           Signals   `json:"signals" bson:"signals"`
	Gates             []string  `json:"gates" bson:"gates"`
	Verdict           string    `json:"verdict" bson:"verdict"`
	Reasons           []string  `json:"reasons" bson:"reasons"`
	Attempt           int       `json:"attempt" bson:"attempt"`
	NeedsManualReview bool      `json:"needs_manual_review" bson:"needs_manual_review"`
	ComputedAt        time.Time `json:"computed_at" bson:"computed_at"`
	PipelineVersion   string    `json:"pipeline_version" bson:"pipeline_version"`
}

// Evaluate applies the stage-1 gates. Every matching gate becomes a reason;
// the verdict is the strongest one (manual > enrich > create). Nothing else
// influences the verdict in stage 1 — no score, no weights.
func Evaluate(s Signals) (verdict string, reasons []string) {
	reasons = []string{}
	manual, enrich := false, false
	// Order matches the spec table: the most diagnostic gates first.
	if s.Kind == KindStub {
		reasons = append(reasons, GateStub)
		enrich = true
	}
	if s.TruncatedByImport {
		reasons = append(reasons, GateTruncated)
		enrich = true
	}
	if s.Words == 0 && !s.HasSourceLink {
		reasons = append(reasons, GateEmpty)
		manual = true
	}
	if s.Mojibake {
		reasons = append(reasons, GateMojibake)
		manual = true
	}
	switch {
	case manual:
		return VerdictManual, reasons
	case enrich:
		return VerdictEnrich, reasons
	default:
		return VerdictCreate, reasons
	}
}

// SameReports whether two assessments of the same source_hash carry identical
// signals — the stop rule "nothing changed since the previous assessment".
func SameSignals(a, b Signals) bool {
	return a.Words == b.Words &&
		a.ProseWords == b.ProseWords &&
		a.EndsWithSentence == b.EndsWithSentence &&
		a.TruncatedByImport == b.TruncatedByImport &&
		a.UnclosedFence == b.UnclosedFence &&
		a.Headings == b.Headings &&
		floatPtrEq(a.CoherenceMin, b.CoherenceMin) &&
		floatPtrEq(a.CoherenceMedian, b.CoherenceMedian) &&
		a.Sentences == b.Sentences &&
		a.ProseShare == b.ProseShare &&
		a.Kind == b.Kind &&
		a.FragmentShare == b.FragmentShare &&
		a.MaxBlockWords == b.MaxBlockWords &&
		a.HasSourceLink == b.HasSourceLink &&
		a.TitleWords == b.TitleWords &&
		a.TitleGeneric == b.TitleGeneric &&
		floatPtrEq(a.TitleTextSimilarity, b.TitleTextSimilarity) &&
		a.Origin == b.Origin &&
		a.Keywords == b.Keywords &&
		a.Links == b.Links &&
		a.HasEmbedding == b.HasEmbedding &&
		a.Mojibake == b.Mojibake
}

func floatPtrEq(a, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
